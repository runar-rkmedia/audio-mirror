package untold

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/runar-rkmedia/audio-mirror/genapi"
	"github.com/runar-rkmedia/audio-mirror/rss"
)

type (
	testCache struct{}
	testHTTP  struct{}
)

func (t *testCache) Retrieve(keyPaths []string, changedAfter time.Time) ([]byte, bool, error) {
	return []byte(testChannelResponse), true, nil
}

func (t *testCache) Write(keyPaths []string, value []byte) (string, error) {
	return "key", nil
}

func (t *testHTTP) Do(req *http.Request) (*http.Response, error) {
	panic("Test dont run http!")
}

func mustParseTime(t *testing.T, s string) *time.Time {
	time, err := time.Parse("2006-01-02T15:04:05Z07:00", s)
	if err != nil {
		t.Fatal(err)
	}
	return &time
}

func TestNewUntoldAPI(t *testing.T) {
	cache := testCache{}
	http_ := testHTTP{}
	options := genapi.GenAPIOptions{
		Logger: slog.Default(),
		Client: &http_,
		Cache:  &cache,
	}
	untold, err := NewUntoldAPI(UntoldAPIOptions{
		GenAPIOptions: options,
		Token:         "footoken",
		Name:          "foobar",
	})
	if err != nil {
		t.Errorf("NewUntoldAPI() error = %v", err)
		return
	}

	want := genapi.GenAPIChannelList{
		Channels: []genapi.GenApiChannel{
			{
				Meta: genapi.GenApiChannelMeta{
					ID:        "sample-id-1",
					Frequency: "Weekly",
					Source:    "foobar",
					SourceURL: "/api/v1/podcasts/original",
					LastAired: mustParseTime(t, "2024-08-01T00:00:00Z"),
				},
				Channel: rss.Channel{
					Locked:      "yes",
					Title:       "Sample Podcast 1",
					Description: "Sample description for the first item. This is where you would describe the content.",
					Author:      "Sample Producer",
					Image: rss.Image{
						URL: "https://example.com/images/sample1_cover_lg.png",
					},
				},
			},
			{
				Meta: genapi.GenApiChannelMeta{
					ID:        "sample-id-2",
					Frequency: "Monthly",
					Source:    "foobar",
					SourceURL: "/api/v1/podcasts/original",
					LastAired: mustParseTime(t, "2024-07-15T00:00:00Z"),
				},
				Channel: rss.Channel{
					Locked:      "yes",
					Title:       "Sample Podcast 2",
					Description: "Sample description for the second item. This is where you would describe the content.",
					Author:      "Sample Producer",
					Image: rss.Image{
						URL: "https://example.com/images/sample2_cover_lg.png",
					},
				},
			},
			{
				Meta: genapi.GenApiChannelMeta{
					ID:        "sample-id-3",
					Frequency: "Daily",
					Source:    "foobar",
					SourceURL: "/api/v1/podcasts/original",
					LastAired: mustParseTime(t, "2024-07-30T00:00:00Z"),
				},
				Channel: rss.Channel{
					Locked:      "yes",
					Title:       "Sample Podcast 3",
					Description: "Sample description for the third item. This is where you would describe the content.",
					Author:      "Sample Producer",
					Image: rss.Image{
						URL: "https://example.com/images/sample3_cover_lg.png",
					},
				},
			},
		},
	}
	ctx := context.TODO()
	var got genapi.GenAPIChannelList
	err = untold.DecodeEndpointData(ctx, *untold.EndpointOriginals, "json", []byte(testChannelResponse), &got.Channels)
	if err != nil {
		t.Errorf("DecodeEndpointData() error = %v", err)
	}
	if diff := deep.Equal(want, got); len(diff) != 0 {
		for _, v := range diff {
			t.Errorf("not equal %v", v)
		}
	}
}

var testChannelResponse = `[
  {
    "banner": {
      "blurhash": "LKNf^p00t7of9FX8V@j?1RY-WYay",
      "lg": "https://example.com/images/sample1_lg.png",
      "md": "https://example.com/images/sample1_md.png",
      "sm": "https://example.com/images/sample1_sm.png"
    },
    "cover": {
      "blurhash": "U8S8G:~V0of00LWBX8j@o#j[Ioay",
      "lg": "https://example.com/images/sample1_cover_lg.png",
      "md": "https://example.com/images/sample1_cover_md.png",
      "sm": "https://example.com/images/sample1_cover_sm.png"
    },
    "description": "Sample description for the first item. This is where you would describe the content.",
    "frequency": "Weekly",
    "full": {
      "blurhash": "ToNF}p000Lof000jWBay00V@ofay",
      "lg": "https://example.com/images/sample1_full_lg.png",
      "md": "https://example.com/images/sample1_full_md.png",
      "sm": "https://example.com/images/sample1_full_sm.png"
    },
    "hero": {
      "blurhash": "TNF~}p00Mof9G~WBV@j@00t7ayof",
      "lg": "https://example.com/images/sample1_hero_lg.png",
      "md": "https://example.com/images/sample1_hero_md.png",
      "sm": "https://example.com/images/sample1_hero_sm.png"
    },
    "id": "sample-id-1",
    "lastEpisodeDate": "2024-08-01T00:00:00Z",
    "name": "Sample Podcast 1",
    "original": true,
    "producer": "Sample Producer"
  },
  {
    "banner": {
      "blurhash": "LKJF~p00V8ay9FX@V@j@00o#WYay",
      "lg": "https://example.com/images/sample2_lg.png",
      "md": "https://example.com/images/sample2_md.png",
      "sm": "https://example.com/images/sample2_sm.png"
    },
    "cover": {
      "blurhash": "U9S8G~@R0of00LWBX8j@o#ayWYay",
      "lg": "https://example.com/images/sample2_cover_lg.png",
      "md": "https://example.com/images/sample2_cover_md.png",
      "sm": "https://example.com/images/sample2_cover_sm.png"
    },
    "description": "Sample description for the second item. This is where you would describe the content.",
    "frequency": "Monthly",
    "full": {
      "blurhash": "ToNF}p00Lof000jWBay00V@ofay",
      "lg": "https://example.com/images/sample2_full_lg.png",
      "md": "https://example.com/images/sample2_full_md.png",
      "sm": "https://example.com/images/sample2_full_sm.png"
    },
    "hero": {
      "blurhash": "T8F~}p00Nof9G~WBV@j@00t7ayof",
      "lg": "https://example.com/images/sample2_hero_lg.png",
      "md": "https://example.com/images/sample2_hero_md.png",
      "sm": "https://example.com/images/sample2_hero_sm.png"
    },
    "id": "sample-id-2",
    "lastEpisodeDate": "2024-07-15T00:00:00Z",
    "name": "Sample Podcast 2",
    "original": true,
    "producer": "Sample Producer"
  },
  {
    "banner": {
      "blurhash": "L8KF~p00W8ay9FX@V@j@00o#WYay",
      "lg": "https://example.com/images/sample3_lg.png",
      "md": "https://example.com/images/sample3_md.png",
      "sm": "https://example.com/images/sample3_sm.png"
    },
    "cover": {
      "blurhash": "U7S8G~@R0of00LWBX8j@o#ayWYay",
      "lg": "https://example.com/images/sample3_cover_lg.png",
      "md": "https://example.com/images/sample3_cover_md.png",
      "sm": "https://example.com/images/sample3_cover_sm.png"
    },
    "description": "Sample description for the third item. This is where you would describe the content.",
    "frequency": "Daily",
    "full": {
      "blurhash": "ToNF}p00Lof000jWBay00V@ofay",
      "lg": "https://example.com/images/sample3_full_lg.png",
      "md": "https://example.com/images/sample3_full_md.png",
      "sm": "https://example.com/images/sample3_full_sm.png"
    },
    "hero": {
      "blurhash": "T7F~}p00Mof9G~WBV@j@00t7ayof",
      "lg": "https://example.com/images/sample3_hero_lg.png",
      "md": "https://example.com/images/sample3_hero_md.png",
      "sm": "https://example.com/images/sample3_hero_sm.png"
    },
    "id": "sample-id-3",
    "lastEpisodeDate": "2024-07-30T00:00:00Z",
    "name": "Sample Podcast 3",
    "original": true,
    "producer": "Sample Producer"
  }
]`
