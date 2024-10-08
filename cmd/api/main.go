package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/runar-rkmedia/audio-mirror/cache"
	"github.com/runar-rkmedia/audio-mirror/db"
	apiv1 "github.com/runar-rkmedia/audio-mirror/gen/api/v1" // generated by protoc-gen-go
	"github.com/runar-rkmedia/audio-mirror/gen/api/v1/apiv1connect"
	"github.com/runar-rkmedia/audio-mirror/genapi"
	untold "github.com/runar-rkmedia/audio-mirror/genapi/apiuntold"
	"github.com/runar-rkmedia/audio-mirror/logger"
	"github.com/runar-rkmedia/audio-mirror/rss"
)

type APIServer struct {
	OriginHost         string
	OriginScheme       string
	TempChannelList    []genapi.GenAPIChannelList
	TempChannelFinders []ChannelFinder
}

// GetEpisodes implements apiv1connect.FeedServiceHandler.
func (s *APIServer) GetEpisodes(ctx context.Context, req *connect.Request[apiv1.GetEpisodesRequest]) (*connect.Response[apiv1.GetEpisodesResponse], error) {
	chlists := genapi.FlattenAndDeduplicate(nil, s.TempChannelList)
	for _, v := range chlists.Channels {
		fmt.Println(req.Msg.Id, "=?", v.Meta.ID)
		if v.Meta.ID != req.Msg.Id {
			continue
		}
		for _, lister := range s.TempChannelFinders {
			episodes, _, err := lister.ListEpisodes(ctx, v.Meta.ID)
			if err != nil {
				return nil, err
			}
			episodesPayload := mapEpsiodes(episodes.Items)
			return &connect.Response[apiv1.GetEpisodesResponse]{
				Msg: &apiv1.GetEpisodesResponse{
					Episodes: episodesPayload,
				},
			}, nil

		}
	}
	return nil, nil
}

func (s *APIServer) getOrigin(req headerProvider) string {
	host := s.OriginHost
	if host == "" {
		host = req.Header().Get("host")
	}
	return s.OriginScheme + host
}

type headerProvider interface {
	Header() http.Header
}

func (s *APIServer) mapChannel(v genapi.GenApiChannel, req headerProvider) *apiv1.Channel {
	apiChannel := &apiv1.Channel{
		Id:          v.GUID,
		Type:        ChanType(v.Meta.Kind),
		Title:       v.Title,
		Description: v.Description,
		ImageUrl:    v.Image.URL,
	}
	if apiChannel.Id == "" {
		apiChannel.Id = v.Meta.ID
	}
	apiChannel.FeedUrl = s.getOrigin(req) + "/" + apiChannel.Id
	return apiChannel
}

func mapEpsiodes(episodes []rss.Item) []*apiv1.Episode {
	episodesPayload := make([]*apiv1.Episode, len(episodes))
	for i, item := range episodes {
		epi := apiv1.Episode{
			Id:          item.GUID,
			ImageUrl:    item.Image.URL,
			Title:       item.Title,
			Description: item.Description,
			SoundUrl:    item.Link,
		}
		episodesPayload[i] = &epi
	}
	return episodesPayload
}

func (s *APIServer) GetChannels(
	ctx context.Context,
	req *connect.Request[apiv1.GetChannelsRequest],
) (*connect.Response[apiv1.GetChannelsResponse], error) {
	res := connect.NewResponse(&apiv1.GetChannelsResponse{
		Channels: []*apiv1.Channel{},
	})
	fmt.Println("source", s.TempChannelList[0].Channels[0].Meta)
	chlists := genapi.FlattenAndDeduplicate(nil, s.TempChannelList)
	for _, v := range chlists.Channels {
		res.Msg.Channels = append(res.Msg.Channels, s.mapChannel(v, req))
	}
	return res, nil
}

func (s *APIServer) GetChannel(
	ctx context.Context,
	req *connect.Request[apiv1.GetChannelRequest],
) (*connect.Response[apiv1.GetChannelResponse], error) {
	for k, v := range req.Header() {
		fmt.Println("header", k, v)
	}
	res := connect.NewResponse(&apiv1.GetChannelResponse{
		Channel:  &apiv1.Channel{},
		Episodes: []*apiv1.Episode{},
	})
	chlists := genapi.FlattenAndDeduplicate(nil, s.TempChannelList)
	for _, v := range chlists.Channels {
		if v.Meta.ID != req.Msg.Id {
			continue
		}
		ch := s.mapChannel(v, req)
		res.Msg.Channel = ch
		for _, lister := range s.TempChannelFinders {
			episodes, _, err := lister.ListEpisodes(ctx, v.Meta.ID)
			if err != nil {
				return nil, err
			}
			episodesPayload := mapEpsiodes(episodes.Items)
			res.Msg.Episodes = episodesPayload
			return res, nil

		}
	}
	return res, nil
}

func ChanType(c genapi.ChannelType) apiv1.ChannelType {
	switch c {
	case genapi.ChannelTypePodCast:
		return apiv1.ChannelType_CHANNEL_TYPE_PODCAST
	case genapi.ChannelTypeAudioBook:
		return apiv1.ChannelType_CHANNEL_TYPE_AUDIO_BOOK
	}
	return apiv1.ChannelType_CHANNEL_TYPE_UNSPECIFIED
}

func main() {
	originHost := flag.String("originhost", "", "Set the host to use. Most proxies does not expose the real host to server, so this can set it manually")
	flag.Parse()
	if *originHost == "" {
		*originHost = os.Getenv("AUDIO_MIRROR_ORIGINHOST")
	}
	l, err := logger.CreateLogger(logger.LogOptions{})
	if err != nil {
		panic("Failed to create logger" + err.Error())
	}
	slog.SetDefault(l.Logger)
	db, err := db.CreateDatabase(db.DBOptions{
		InMemory: false,
		FilePath: "./db.sqlite3",
	})
	if err != nil {
		l.FatalErr("failed to create database", err)
	}
	ctx := context.TODO()
	db.GetChannels(ctx)
	untold, err := initUntold(l)
	if err != nil {
		l.FatalErr("failed to init untold", err)
	}
	// Temp
	channelLists, err := FindChannels(context.TODO(), l, untold)
	if err != nil {
		l.FatalErr("Failed to Find channels", err)
	}
	feedServer := &APIServer{TempChannelList: channelLists, TempChannelFinders: []ChannelFinder{untold}, OriginHost: *originHost}

	switch feedServer.OriginScheme {
	case "":
		feedServer.OriginScheme = "https://"
	case "https":
		feedServer.OriginScheme = "https://"
	case "http":
		feedServer.OriginScheme = "http://"
	}
	mux := http.NewServeMux()
	path, handler := apiv1connect.NewFeedServiceHandler(feedServer)
	mux.Handle(path, handler)
	mux.HandleFunc("GET /feed/{id}", feedServer.HandleRssFeed)
	mux.HandleFunc("/", proxyPass)
	address := "0.0.0.0:8080"
	l.Info("Webserver starting", slog.String("address", address), slog.String("originHost", feedServer.OriginHost))

	err = http.ListenAndServe(
		address,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		l.FatalErr("Failed to create listener", err)
	}
}

func proxyPass(res http.ResponseWriter, req *http.Request) {
	// Encrypt Request here
	// ...

	url, _ := url.Parse("http://localhost:5173")
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(res, req)
}

func (s *APIServer) HandleRssFeed(w http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("id")
	fmt.Println("incoming", idString)
	chlists := genapi.FlattenAndDeduplicate(nil, s.TempChannelList)
	for _, channel := range chlists.Channels {
		if channel.GUID == idString || channel.Meta.ID == idString {
			rssFeed := rss.RssHeader(channel.Channel)
			x, err := xml.MarshalIndent(rssFeed, "", "  ")
			fmt.Println("channel", channel.Title, len(channel.Item))
			w.Header().Set("Content-Type", "application/xml")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			// return
			if _, err := w.Write(x); err != nil {
				w.WriteHeader(500)
				return
			}
			return
		}
	}
	w.WriteHeader(400)
}

type ChannelFinder interface {
	FindAllChannels(ctx context.Context) ([]genapi.GenAPIChannelList, error)
	ListEpisodes(ctx context.Context, id string) (*genapi.GenAPIEpisodeList, *http.Response, error)
}

// deprecated only here temproarily during development until there is a database
func initUntold(l *logger.Logger) (ChannelFinder, error) {
	client := http.Client{}
	cacheDir := "./.cache"
	cacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		l.Fatal("Failed to get absolute path for cacheDir", slog.String("cacheDir", cacheDir), slog.Any("error", err))
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		l.Fatal("Failed to initiate cache for cacheDir", slog.String("cacheDir", cacheDir), slog.Any("error", err))
	}
	cache := cache.NewCache(cacheDir)
	genOptions := genapi.GenAPIOptions{
		Logger: l.Logger,
		Client: &client,
		Cache:  cache,
	}
	untoldToken := os.Getenv("UNTOLD_TOKEN")
	if untoldToken == "" {
		l.Fatal("UNTOLD_TOKEN not set, quitting")
	}
	options := untold.UntoldAPIOptions{
		GenAPIOptions: genOptions,
		Token:         untoldToken,
	}
	untold, err := untold.NewUntoldAPI(options)
	return untold, err
}

// deprecated will not be used to find episodes in the future.
func FindChannels(ctx context.Context, l *logger.Logger, channelFinder ChannelFinder) ([]genapi.GenAPIChannelList, error) {
	channelLists, err := channelFinder.FindAllChannels(ctx)
	if err != nil {
		return channelLists, err
	}
	for _, lists := range channelLists {
		for i := range lists.Channels {
			channel := &lists.Channels[i]
			episodes, _, err := channelFinder.ListEpisodes(ctx, channel.Meta.ID)
			if err != nil {
				l.Fatal("failed?", slog.Any("error", err))
			}
			channel.Item = episodes.Items
		}
	}
	return channelLists, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	dur := fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	fmt.Println("duration", dur, d)
	return dur
}
