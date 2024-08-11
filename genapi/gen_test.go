package genapi

import (
	"context"
	"log/slog"
	"net/url"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/runar-rkmedia/audio-mirror/rss"
)

func TestGenAPI_CreateSubURL(t *testing.T) {
	testURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	testEndpoint := Endpoint{URL: testURL}
	tests := []struct {
		name     string
		base     Endpoint
		endpoint GenAPIEndpoint
		data     map[string]any
		want     string
		wantErr  bool
	}{
		{
			"Should return valid url for template in url",
			testEndpoint,
			GenAPIEndpoint{Path: "/foo/{{.ID}}/bar"},
			map[string]any{"ID": "123abc"},
			"https://example.com/foo/123abc/bar",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GenAPI{
				Endpoint: tt.base,
				GenAPIOptions: GenAPIOptions{
					Logger: slog.Default(),
				},
			}
			got, err := g.CreateSubURL(g.URL, tt.endpoint, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAPI.CreateSubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("GenAPI.CreateSubURL() = \n'%v', want \n'%v'", got.String(), tt.want)
			}
		})
	}
}

func TestGenAPI_DecodeEndpointData(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   GenAPIEndpoint
		dataType   string
		data       []byte
		out        GenAPIChannelList
		wantErr    bool
		wantResult GenAPIChannelList
	}{
		{
			"Should map fields correctlly for single item, single key",
			GenAPIEndpoint{
				Mapping: map[string]string{
					"Title": "foo",
				},
			},
			"json",
			[]byte(`[{"foo": "bar"}]`),
			GenAPIChannelList{},
			false,
			GenAPIChannelList{
				Channels: []GenApiChannel{
					{
						Channel: rss.Channel{
							Title: "bar",
						},
					},
				},
			},
		},
		{
			"Should map fields correctlly for single item, multiple keys",
			GenAPIEndpoint{
				Mapping: map[string]string{
					"Title":       "foo",
					"Description": "someKey",
				},
			},
			"json",
			[]byte(`[{"foo": "bar", "someKey": "Some text"}]`),
			GenAPIChannelList{},
			false,
			GenAPIChannelList{
				Channels: []GenApiChannel{
					{
						Channel: rss.Channel{
							Title:       "bar",
							Description: "Some text",
						},
					},
				},
			},
		},
		{
			"Should map fields correctlly for single item, multiple keys, not on root",
			GenAPIEndpoint{
				RootMapping: "items",
				Mapping: map[string]string{
					"Title":       "foo",
					"Description": "someKey",
				},
			},
			"json",
			[]byte(`{"items": [{"foo": "bar", "someKey": "Some text"}]}`),
			GenAPIChannelList{},
			false,
			GenAPIChannelList{
				Channels: []GenApiChannel{
					{
						Channel: rss.Channel{
							Title:       "bar",
							Description: "Some text",
						},
					},
				},
			},
		},
		{
			"Should map fields correctlly for two values, single key",
			GenAPIEndpoint{
				Mapping: map[string]string{
					"Title": "foo",
				},
			},
			"json",
			[]byte(`[{"foo": "bar"}, {"foo", "baz"}]`),
			GenAPIChannelList{},
			false,
			GenAPIChannelList{
				Channels: []GenApiChannel{
					{
						Channel: rss.Channel{
							Title: "bar",
						},
					},
					{
						Channel: rss.Channel{
							Title: "baz",
						},
					},
				},
			},
		},
		{
			"Should map fields to various types",
			GenAPIEndpoint{
				Mapping: map[string]string{
					"Title":    "foo",
					"Category": "nested.categories|@categories",
				},
			},
			"json",
			[]byte(`[{"foo": "bar", "nested": {"categories": ["a", "b", "c"]}}]`),
			GenAPIChannelList{},
			false,
			GenAPIChannelList{
				Channels: []GenApiChannel{
					{
						Channel: rss.Channel{
							Title: "bar", Category: []rss.Category{
								{AttrText: "a"},
								{AttrText: "b"},
								{AttrText: "c"},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GenAPI{}
			ctx := context.TODO()
			if err := g.DecodeEndpointData(ctx, tt.endpoint, tt.dataType, tt.data, &tt.out.Channels); (err != nil) != tt.wantErr {
				t.Fatalf("GenAPI.DecodeEndpointData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := deep.Equal(tt.wantResult, tt.out); len(diff) != 0 {
				t.Fatalf("not equal %v", diff)
			}
		})
	}
}
