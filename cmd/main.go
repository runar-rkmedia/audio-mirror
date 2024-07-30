package main

import (
	"context"
	"encoding/xml"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/kennygrant/sanitize"
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
	podcasts, errs := untold.FindAllPodcasts(ctx)
	if errs != nil {
		untold.Logger.Error("failed to find all pordcasts", slog.Any("error", err))
	}
	os.MkdirAll("./feeds", 0755)
	if *query != "" {
		searchResult, _, _ := untold.SearchTitles(ctx, *query)
		// for _, podCast := range searchResult {
		// 	fmt.Println(podCast.Name)
		// }
		podcasts = append(podcasts, searchResult...)
		for _, podCast := range podcasts {
			l := untold.Logger.With(
				slog.String("podcast-name", podCast.Name),
			)
			l.Info("Fetching episodes for podcast")
			episodes, _, _ := untold.ListEpisodes(ctx, podCast.ID)
			rss, err := CreateUntoldRss(ctx, podCast, episodes)
			if err != nil {
				l.Error("Failed to create rss-feed", slog.Any("error", err))
				break
			}
			l = untold.Logger.With(
				slog.Int("episode-count", len(episodes)),
			)
			x, err := xml.MarshalIndent(rss, "", "  ")
			if err != nil {
				l.Error("Failed to marshal rss-feed", slog.Any("error", err))
				break
			}
			sane := sanitize.BaseName(podCast.Name)
			filePath := path.Join("./feeds", sane)
			l = l.With(slog.String("filepath", filePath), slog.Int("size", len(x)))
			err = os.WriteFile(filePath, x, 0644)
			if err != nil {
				l.Error("Failed to write rss-file", slog.Any("error", err))
				break
			}
			l.Debug("Wrote rss-file")
		}
	}
}
