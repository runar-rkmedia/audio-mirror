package genapi

import (
	"log/slog"
	"net/url"
	"reflect"
	"testing"
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
