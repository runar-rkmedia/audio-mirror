package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/runar-rkmedia/audio-mirror/genapi"
)

type UntoldAPI struct {
	*genapi.GenAPI
	EndpointOriginals, EndpointFollowing, EndpointRecommended, EndpointHero, EndpointCategoriesDiscover, EndpointPopularEpisodes, EndpointFollowedEpisodes *genapi.GenAPIEndpoint
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
	// Their search is a bit strange. Not all podcasts seems searchable, for instance Truecrimepodden, which does show up in categories/discover
	untold.EndpointSearchTitles = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/search", Query: "term={{.query}}"}
	untold.EndpointListEpisodes = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/{{.podID}}/episodes"}
	untold.EndpointOriginals = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/original"}
	untold.EndpointFollowing = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/followed"}
	untold.EndpointRecommended = &genapi.GenAPIEndpoint{Path: "/api/v1/podcasts/recommended"}
	untold.EndpointHero = &genapi.GenAPIEndpoint{Path: "/api/v1/hero"}
	untold.EndpointCategories = &genapi.GenAPIEndpoint{Path: "/api/v1/discover"}
	// These will  be added to search-results in the app.
	untold.EndpointCategoriesDiscover = &genapi.GenAPIEndpoint{Path: "/api/v1/categories/discover"}
	untold.EndpointPopularEpisodes = &genapi.GenAPIEndpoint{Path: "/api/v1/episodes/popular"}
	untold.EndpointFollowedEpisodes = &genapi.GenAPIEndpoint{Path: "/api/v1/episodes/followed"}

	return untold, nil
}

// Untold api is a bit weird. As far as I can tell, there is no endpoint for simply listing all podcasts.
// Probably this is because there are third-party-podcasts here as well.
// However, there are a few endpoints that seem to list the most interesting, and exclusive episodes.
func (u *UntoldAPI) FindAllPodcasts(ctx context.Context) ([]UntoldPodcast, error) {
	var errs error
	var j []UntoldPodcast
	{
		o, _, err := u.GetOriginals(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			j = append(j, o...)
		}
	}
	{
		o, _, err := u.GetRecommended(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			j = append(j, o...)
		}
	}
	{
		o, _, err := u.GetFollowed(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			for _, v := range o {
				j = append(j, v.UntoldPodcast)
			}
		}
	}
	{
		o, _, err := u.GetCategoriesDiscover(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			for _, v := range o {
				j = append(j, v.Podcasts...)
			}
		}
	}
	{
		o, _, err := u.GetHero(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			for _, v := range o {
				j = append(j, v.Podcast)
			}
		}
	}

	return j, errs
}
func (u *UntoldAPI) GetRecommended(ctx context.Context) ([]UntoldPodcast, *http.Response, error) {
	var j []UntoldPodcast
	r, err := u.RunEndpoint(ctx, *u.EndpointRecommended, nil, "recommended", &j)
	u.Logger.Debug("Got recommended", slog.Int("count", len(j)))
	return j, r, err
}
func (u *UntoldAPI) GetHero(ctx context.Context) ([]UntoldHero, *http.Response, error) {
	var j []UntoldHero
	r, err := u.RunEndpoint(ctx, *u.EndpointHero, nil, "hero", &j)
	u.Logger.Debug("Got hero", slog.Int("count", len(j)))
	return j, r, err
}
func (u *UntoldAPI) GetCategoriesDiscover(ctx context.Context) ([]UntoldCategoryDiscoverElement, *http.Response, error) {
	var j []UntoldCategoryDiscoverElement
	r, err := u.RunEndpoint(ctx, *u.EndpointCategoriesDiscover, nil, "categories-discover", &j)
	u.Logger.Debug("Got result of discover for categories", slog.Int("count", len(j)))
	return j, r, err
}
func (u *UntoldAPI) GetPopularEpisodes(ctx context.Context) ([]UntoldPodcastWithEpisode, *http.Response, error) {
	var j []UntoldPodcastWithEpisode
	r, err := u.RunEndpoint(ctx, *u.EndpointPopularEpisodes, nil, "popular-episodes", &j)
	u.Logger.Debug("Got popular-episodes", slog.Int("count", len(j)))
	return j, r, err
}
func (u *UntoldAPI) GetFollowedEpisodes(ctx context.Context) ([]UntoldPodcastWithEpisode, *http.Response, error) {
	var j []UntoldPodcastWithEpisode
	r, err := u.RunEndpoint(ctx, *u.EndpointFollowedEpisodes, nil, "followed-episodes", &j)
	u.Logger.Debug("Got followed-episodes", slog.Int("count", len(j)))
	return j, r, err
}

// Returns some data about the file. Mostly, just the content-type is of direct value
func (u *UntoldAPI) GetMediaInfo(ctx context.Context, soundUrl string) (*http.Response, error) {
	r, err := u.NewRequest(ctx, http.MethodGet, soundUrl, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("range", "bytes=0-1")
	return u.Client.Do(r)
}

// Returns a list of categories, even though they call their endpoint discover
func (u *UntoldAPI) GetCategories(ctx context.Context) ([]UntoldDiscoverElement, *http.Response, error) {
	var j []UntoldDiscoverElement
	r, err := u.RunEndpoint(ctx, *u.EndpointCategories, nil, "categories", &j)
	u.Logger.Debug("Got categories (discover)", slog.Int("count", len(j)))
	return j, r, err
}

func (u *UntoldAPI) ListEpisodes(ctx context.Context, podCastID string) ([]UntoldEpisode, *http.Response, error) {
	var j []UntoldEpisode
	r, err := u.RunEndpoint(ctx, *u.EndpointListEpisodes, map[string]any{"podID": podCastID}, "episodes-", &j)
	u.Logger.Debug("Got episodes", slog.Int("count", len(j)))
	return j, r, err
}

func (u *UntoldAPI) SearchTitles(ctx context.Context, query string) ([]UntoldPodcast, *http.Response, error) {
	var j []UntoldPodcast
	r, err := u.RunEndpoint(ctx, *u.EndpointSearchTitles, map[string]any{"query": query}, "searchx", &j)
	u.Logger.Debug("Got Searchresults", slog.Int("count", len(j)))
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
type UntoldPodcastWithEpisode struct {
	Episode UntoldEpisode `json:"episode"`
	Podcast UntoldPodcast `json:"podcast"`
}

type UntoldCategoryDiscoverElement struct {
	Format       Format          `json:"format"`
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Podcasts     []UntoldPodcast `json:"podcasts"`
	ShowInSearch bool            `json:"showInSearch"`
}
type Format string

const (
	FormatBanner Format = "banner"
	Cover        Format = "cover"
)

type UntoldHero struct {
	Kicker  *string       `json:"kicker,omitempty"`
	Podcast UntoldPodcast `json:"podcast"`
	Title   string        `json:"title"`
	Type    string        `json:"type"`
}

type UntoldPodcast struct {
	Banner          Image      `json:"banner"`
	Cover           Image      `json:"cover"`
	Description     string     `json:"description"`
	Frequency       *string    `json:"frequency,omitempty"`
	Full            Image      `json:"full"`
	Hero            Image      `json:"hero"`
	ID              string     `json:"id"`
	LastEpisodeDate *time.Time `json:"lastEpisodeDate"`
	Name            string     `json:"name"`
	Original        bool       `json:"original"`
	Producer        string     `json:"producer"`
}

type Image struct {
	Blurhash *string `json:"blurhash,omitempty"`
	Lg       string  `json:"lg"`
	Md       string  `json:"md"`
	Sm       string  `json:"sm"`
}

type UntoldEpisode struct {
	Author      Author        `json:"author"`
	Cover       *Image        `json:"cover"`
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

type UntoldDiscoverElement struct {
	CategoryID *string `json:"categoryId,omitempty"`
	Type       Type    `json:"type"`
}

type Type string

const (
	Category        Type = "category"
	PopularEpisodes Type = "popular_episodes"
	PremierLeague   Type = "premier_league"
)
