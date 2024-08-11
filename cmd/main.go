package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mattn/go-colorable"
	"github.com/runar-rkmedia/audio-mirror/cache"
	"github.com/runar-rkmedia/audio-mirror/genapi"
	untold "github.com/runar-rkmedia/audio-mirror/genapi/apiuntold"
	"hypera.dev/lib/slog/pretty"
)

func Fatal(msg string, args ...any) {
	slog.Default().Error(msg, args...)
	os.Exit(1)
}

func main() {
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
	options := untold.UntoldAPIOptions{
		GenAPIOptions: genOptions,
		Token:         untoldToken,
	}
	untold, err := untold.NewUntoldAPI(options)
	if err != nil {
		Fatal("failed to create general api", slog.Any("error", err))
	}
	podcasts, errs := untold.FindAllChannels(ctx)
	if errs != nil {
		untold.Logger.Error("failed to find all pordcasts", slog.Any("error", err))
	}
	os.MkdirAll("./feeds", 0755)
	for _, v := range podcasts {
		for _, rsschannel := range v.Channels {
			XML, err := xml.MarshalIndent(rsschannel.Channel, "", "  ")
			if err != nil {
				untold.Logger.Error("Failed to write xml for feed", slog.Any("error", err))
			}
			fmt.Println(string(XML))
		}
	}
}
