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
	var podcasts []UntoldPodcast
	// p, _, _ := untold.ListEpisodes(ctx, "3d0429f8-8a04-432b-973f-2c6556467891")
	p, _, _ := untold.GetFollowed(ctx)
	for _, v := range p {
		podcasts = append(podcasts, v.UntoldPodcast)
	}
	os.MkdirAll("./feeds", 0644)
	if *query != "" {
		searchResult, _, _ := untold.SearchTitles(ctx, *query)
		// for _, podCast := range searchResult {
		// 	fmt.Println(podCast.Name)
		// }
		podcasts = append(podcasts, searchResult...)
		for _, podCast := range podcasts {
			untold.Logger.Info("Fetching episodes for podcast", slog.String("name", podCast.Name))
			episodes, _, _ := untold.ListEpisodes(ctx, podCast.ID)
			rss, err := CreateUntoldRss(ctx, podCast, episodes)
			if err != nil {
				untold.Logger.Error("Failed to create rss-feed", slog.Any("error", err))
				break
			}
			x, err := xml.MarshalIndent(rss, "", "  ")
			if err != nil {
				untold.Logger.Error("Failed to marshal rss-feed", slog.Any("error", err))
				break
			}
			sane := sanitize.BaseName(podCast.Name)
			os.WriteFile(path.Join("./feeds", sane), x, 0644)

		}
	}
}
