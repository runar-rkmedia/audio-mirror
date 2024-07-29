package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/runar-rkmedia/audio-mirror/cache"
	"github.com/runar-rkmedia/audio-mirror/genapi"
	"hypera.dev/lib/slog/pretty"
)

func Fatal(msg string, args ...any) {
	slog.Default().Error(msg, args...)
	os.Exit(1)
}

func main() {
	query := flag.String("query", "", "Set to search")
	flag.Parse()
	opts := &pretty.Options{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	l := slog.New(pretty.NewHandler(colorable.NewColorable(os.Stderr), opts))
	ctx := context.TODO()
	client := http.Client{}
	cacheDir := "./.cache"
	cacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		Fatal("Failed to get absolute path for cacheDir", slog.String("cacheDir", cacheDir), slog.Any("error", err))
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		Fatal("Failed to initiate cache for cacheDir", slog.String("cacheDir", cacheDir), slog.Any("error", err))
	}
	cache := cache.NewCache(cacheDir)

	genOptions := genapi.GenAPIOptions{
		Logger: l,
		Client: &client,
		Cache:  cache,
	}
	untoldToken := os.Getenv("UNTOLD_TOKEN")
	if untoldToken == "" {
		Fatal("UNTOLD_TOKEN not set, quitting")
	}
	untold, err := NewUntoldAPI(untoldToken, genOptions)
	if err != nil {
		Fatal("failed to create general api", slog.Any("error", err))
	}
	if *query != "" {
		searchResult, _, _ := untold.SearchTitles(ctx, *query)
		for _, v := range searchResult {
			untold.Logger.Info("Fetching episodes for podcast", slog.String("name", v.Name))
			untold.ListEpisodes(ctx, v.ID)
		}
	}
}

type UntoldAPI struct {
	*genapi.GenAPI
	EndpointOriginals, EndpointFollowing, EndpointRecommended *genapi.GenAPIEndpoint
}

func NewUntoldAPI(token string, options genapi.GenAPIOptions) (*UntoldAPI, error) {
	untoldHeaders := map[string]string{
		"accept":          "*/*",
		"authorization":   token,
		"x-app-os":        "ios",
		"user-agent":      "Untold/2 CFNetwork/1496.0.7 Darwin/23.5.0",
		"accept-language": "nb-NO,nb;q=0.9",
		"x-app-version":   "1.8.2",
	}
	untoldEndpoint, err := genapi.NewEndoint(&url.URL{
		Scheme: "https",
		Host:   "api.fole.app.iterate.no",
	}, untoldHeaders)
	if err != nil {
		return nil, err
	}
	api, err := genapi.NewGeneralAPI(
		"untold",
		untoldEndpoint,
		options,
	)
	if err != nil {
		return nil, err
	}

	untold := &UntoldAPI{
		GenAPI: api,
	}
	untold.EndpointSearchTitles = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/search", Query: "term={{.query}}"}
	untold.EndpointListEpisodes = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/{{.podID}}/episodes"}
	untold.EndpointOriginals = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/original"}
	untold.EndpointFollowing = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/followed"}
	untold.EndpointRecommended = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/recommended"}

	return untold, nil
}

func (u *UntoldAPI) GetRecommended(ctx context.Context) ([]UntoldPodcast, *http.Response, error) {
	var j []UntoldPodcast
	r, err := u.RunEndpoint(ctx, *u.EndpointRecommended, nil, "recommended", &j)
	u.Logger.Debug("Got recommended", slog.Int("count", len(j)))
	return j, r, err
}

func (u *UntoldAPI) ListEpisodes(ctx context.Context, podCastID string) ([]UntoldEpisode, *http.Response, error) {
	var j []UntoldEpisode
	r, err := u.RunEndpoint(ctx, *u.EndpointListEpisodes, map[string]any{"podID": podCastID}, "episodes-", &j)
	u.Logger.Debug("Got episodes", slog.Int("count", len(j)))
	return j, r, err
}

func (u *UntoldAPI) GetFollowed(ctx context.Context) ([]UntoldFollowed, *http.Response, error) {
	var j []UntoldFollowed
	r, err := u.RunEndpoint(ctx, *u.EndpointFollowing, nil, "followed", &j)
	u.Logger.Debug("Got followed", slog.Int("count", len(j)))
	return j, r, err
}

func (u *UntoldAPI) GetOriginals(ctx context.Context) ([]UntoldPodcast, *http.Response, error) {
	var j []UntoldPodcast
	r, err := u.RunEndpoint(ctx, *u.EndpointOriginals, nil, "originals", &j)
	u.Logger.Debug("Got originals", slog.Int("count", len(j)))
	return j, r, err
}

type UntoldFollowed struct {
	EpisodesPublishedAfterFollow []string `json:"episodesPublishedAfterFollow"`
	UntoldPodcast                `json:"podcast"`
}

type UntoldPodcast struct {
	Banner          Banner     `json:"banner"`
	Cover           Banner     `json:"cover"`
	Description     string     `json:"description"`
	Frequency       *string    `json:"frequency,omitempty"`
	Full            Banner     `json:"full"`
	Hero            Banner     `json:"hero"`
	ID              string     `json:"id"`
	LastEpisodeDate *time.Time `json:"lastEpisodeDate"`
	Name            string     `json:"name"`
	Original        bool       `json:"original"`
	Producer        string     `json:"producer"`
}

type Banner struct {
	Blurhash *string `json:"blurhash,omitempty"`
	Lg       string  `json:"lg"`
	Md       string  `json:"md"`
	Sm       string  `json:"sm"`
}

type UntoldEpisode struct {
	Author      Author        `json:"author"`
	Cover       *Banner       `json:"cover"`
	Description string        `json:"description"`
	Duration    int64         `json:"duration"`
	Full        interface{}   `json:"full"`
	ID          string        `json:"id"`
	Original    bool          `json:"original"`
	Permissions []interface{} `json:"permissions"`
	PodcastID   string        `json:"podcastId"`
	Published   time.Time     `json:"published"`
	Season      Season        `json:"season"`
	SoundURL    string        `json:"soundUrl"`
	Title       string        `json:"title"`
}

type Author string

type Season string
