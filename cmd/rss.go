package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/runar-rkmedia/audio-mirror/rss"
)

func CreateUntoldRss(ctx context.Context, podcast UntoldPodcast, episodes []UntoldEpisode) (rss.Rss, error) {
	var mainImage Banner
	switch {
	case podcast.Cover.Lg != "":
		mainImage = podcast.Cover
	case podcast.Full.Lg != "":
		mainImage = podcast.Full
	case podcast.Hero.Lg != "":
		mainImage = podcast.Hero
	case podcast.Banner.Lg != "":
		mainImage = podcast.Banner
	}
	feed := RssHeader(rss.Channel{
		Link:           rss.Link{},
		Language:       "nb-NO",
		Copyright:      "",
		WebMaster:      "",
		ManagingEditor: "",
		Image: rss.Image{
			// Href:  mainImage.Lg,
			URL:   mainImage.Lg,
			Title: podcast.Name,
		},
		Owner: rss.Owner{
			Name:  "Untold",
			Email: "",
		},
		Keywords:      "",
		Explicit:      "",
		PubDate:       "",
		Title:         podcast.Name,
		Author:        podcast.Producer,
		Description:   podcast.Description,
		Summary:       "",
		Subtitle:      "",
		LastBuildDate: "",
		Item:          make([]rss.Item, len(episodes)),
	})

	for i, epi := range episodes {
		feed.Channel.Item[i] = rss.Item{
			Title:       epi.Title,
			Description: epi.Description,
			Summary:     "",
			Subtitle:    "",
			Category:    rss.ItemCategory{},
			Enclosure: rss.Enclosure{
				URL:    epi.SoundURL,
				Type:   "audio/mpeg",
				Length: strconv.FormatInt(epi.Duration*1000, 10),
			},
			GUID:     epi.ID,
			Duration: formatDuration(time.Second * time.Duration(epi.Duration)),
			PubDate:  epi.Published.String(),
			Link:     epi.SoundURL,
		}
	}

	return feed, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func RssHeader(channel rss.Channel) rss.Rss {
	// I have not checked what any of these fields actually mean
	return rss.Rss{
		XMLName:    xml.Name{Space: "space", Local: "local"},
		Atom:       "http://www.w3.org/2005/Atom",
		Itunes:     "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Itunesu:    "http://www.itunesu.com/feed",
		Googleplay: "http://www.google.com/schemas/play-podcasts/1.0",
		Version:    "2.0",
		Channel:    channel,
	}
}
