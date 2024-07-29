package genapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrMissingEndpoint = errors.New("missing endpoint")
)

type (
	Cache interface {
		Retrieve(keyPaths []string, changedAfter time.Time) ([]byte, bool, error)
		Write(keyPaths []string, value []byte) (string, error)
	}
	Endpoint struct {
		URL     *url.URL
		Headers map[string]string
	}

	GenAPIEndpoint struct {
		Path   string
		Query  string
		Method string
	}
	GenAPI struct {
		Name string
		Endpoint
		GenAPIOptions
		EndpointSearchTitles *GenAPIEndpoint
		EndpointCategories   *GenAPIEndpoint
		EndpointListEpisodes *GenAPIEndpoint
	}
	GenAPIOptions struct {
		Logger *slog.Logger
		Client *http.Client
		Cache  Cache
	}
	GenAPITitle struct {
		g   *GenAPI
		raw map[string]any
	}
)

type GenAPIEpisodeList struct {
	g   *GenAPI
	raw any
}

func NewEndoint(url *url.URL, headers map[string]string) (Endpoint, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	return Endpoint{
		URL:     url,
		Headers: headers,
	}, nil
}

func NewGeneralAPI(name string, baseEndpoint Endpoint, options GenAPIOptions) (*GenAPI, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if baseEndpoint.URL.Scheme == "" {
		baseEndpoint.URL.Scheme = "https"
	}

	return &GenAPI{
		Name:          name,
		Endpoint:      baseEndpoint,
		GenAPIOptions: options,
	}, nil
}

func (g *GenAPI) SetCache(cache Cache) {
	g.Cache = cache
}

func (g *GenAPI) NewRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	if method == "" {
		if body == nil {
			method = "GET"
		} else {
			method = "POST"
		}
	}
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return r, err
	}
	for k, v := range g.Headers {
		r.Header.Add(k, v)
	}
	return r, nil
}

func (g *GenAPI) NewJSONRequest(ctx context.Context, r *http.Request, v any) (*http.Response, error) {
	l := g.Logger.With(
		slog.String("uri", r.URL.String()),
		slog.String("method", r.Method),
	)
	l.Debug("Performing request")
	res, err := g.Client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()
	contentType := res.Header.Get("Content-Type")
	l = l.With(
		slog.String("contentType", contentType),
		slog.Int("status", res.StatusCode),
		slog.Int64("contentLength", res.ContentLength),
	)
	l.Debug("Got response")
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	if l.Enabled(ctx, -10) {
		l.Debug("Raw body", slog.String("body", string(body)), slog.Any("headers", res.Header))
	}
	if res.StatusCode >= 400 {
		l.Error("unsuccessful statuscode", slog.String("body", string(body)))
		return nil, fmt.Errorf("unsuccessful status-code: %d", res.StatusCode)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		l.Debug("Raw body", slog.String("body", string(body)), slog.Any("error", err))
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	return res, nil
}

func (g *GenAPI) ListTitles(ctx context.Context) ([]GenAPITitle, error) {
	return nil, fmt.Errorf("%w: ListTitles", ErrNotImplemented)
}

func (g *GenAPI) CreateSubURL(u *url.URL, endpoint GenAPIEndpoint, data map[string]any) (*url.URL, error) {
	uri := new(url.URL)
	*uri = *u

	if endpoint.Path != "" {
		uri.Path = strings.TrimSuffix(uri.Path, "/") + "/" + strings.TrimPrefix(endpoint.Path, "/")
	}
	data["endpoint"] = endpoint
	path, err := g.TemplateString(uri.Path, data)
	if err != nil {
		return uri, err
	}
	uri.Path = path
	if uri.RawQuery != "" {

		query, err := g.TemplateString(uri.RawQuery, data)
		if err != nil {
			return uri, err
		}
		uri.RawQuery = query
	}
	if endpoint.Query != "" {
		q := uri.Query()
		query, err := g.TemplateString(endpoint.Query, data)
		if err != nil {
			return uri, err
		}

		qStr, err := url.ParseQuery(query)
		if err != nil {
			return uri, err
		}
		for k, v := range qStr {
			for _, s := range v {
				q.Add(k, s)
			}
		}
		uri.RawQuery = q.Encode()
	}

	return uri, nil
}

func (g *GenAPI) TemplateString(templateString string, vars map[string]any) (string, error) {
	tmpl, err := template.New("_").Parse(templateString)
	if err != nil {
		return templateString, fmt.Errorf("failed to create template from string '%s': %s", templateString, err)
	}
	buf := bytes.NewBufferString("")
	vars["g"] = g
	if g.Logger.Enabled(context.TODO(), slog.LevelDebug) {
		g.Logger.Debug("Running template",
			slog.String("templateString", templateString),
			slog.Any("vars", vars),
		)
	}
	err = tmpl.Execute(buf, vars)
	if err != nil {
		g.Logger.Error("Failed to run template",
			slog.String("templateString", templateString),
			slog.Any("vars", vars),
		)
	}
	return buf.String(), err
}

func (g *GenAPI) getCache(keyPath string) ([]byte, bool) {
	if g.Cache == nil {
		return nil, false
	}
	if filepath.Ext(keyPath) == "" {
		keyPath += ".json"
	}
	keyPaths := []string{
		g.Name,
		keyPath,
	}
	data, ok, err := g.Cache.Retrieve(keyPaths, time.Now().Add(-1*time.Hour))
	if err != nil {
		g.Logger.Error("Failed to get cached item", slog.Any("error", err))
		return nil, false
	}
	return data, ok
}

func (g *GenAPI) writeCache(keyPath string, value any) (string, error) {
	if g.Cache == nil {
		return "", nil
	}
	if filepath.Ext(keyPath) == "" {
		keyPath += ".json"
	}
	keyPaths := []string{
		g.Name,
		keyPath,
	}
	b, err := json.Marshal(value)
	if err != nil {
		g.Logger.Error("Failed to marshall cached item", slog.Any("error", err))
		return "", err
	}
	cacheID, err := g.Cache.Write(keyPaths, b)
	if err != nil {
		g.Logger.Error("Failed to write cached item", slog.Any("error", err))
		return cacheID, err
	}
	return cacheID, err
}

func (g *GenAPI) ListEpisodes(ctx context.Context, id string) (*GenAPIEpisodeList, *http.Response, error) {
	if g.EndpointListEpisodes == nil {
		return nil, nil, fmt.Errorf("%w: ListEpisodesEndpoint", ErrMissingEndpoint)
	}
	data := map[string]any{
		"podID": id,
	}
	endpoint := g.EndpointListEpisodes
	list := &GenAPIEpisodeList{
		g:   g,
		raw: map[string]any{},
	}

	r, err := g.RunEndpoint(ctx, *endpoint, data, "episodes-", &list.raw)
	return list, r, err
}

func (g *GenAPI) SearchTitles(ctx context.Context, query string) ([]struct{ Name, ID string }, *http.Response, error) {
	if g.EndpointSearchTitles == nil {
		return nil, nil, fmt.Errorf("%w: SearchTitlesEndpoint", ErrMissingEndpoint)
	}
	data := map[string]any{
		"query": query,
	}
	endpoint := g.EndpointSearchTitles
	var j []struct {
		Name, ID string
	}
	r, err := g.RunEndpoint(ctx, *endpoint, data, "search-", &j)
	return j, r, err
}

func (g *GenAPI) RunEndpoint(
	ctx context.Context,
	endpoint GenAPIEndpoint,
	data map[string]any,
	cacheKeyPrefix string,
	jsonData any,
) (*http.Response, error) {
	cacheKey := cacheKeyPrefix
	if data == nil {
		data = map[string]any{}
	}
	for k, v := range data {
		cacheKey += k + "=" + fmt.Sprintf("%v", v)
	}
	url, err := g.CreateSubURL(g.URL, endpoint, data)
	if err != nil {
		return nil, err
	}
	cached, found := g.getCache(cacheKey)
	if found && len(cached) > 0 {
		g.Logger.Debug("using cache",
			slog.String("cacheKey", cacheKey),
		)
		err = json.Unmarshal(cached, jsonData)
		if err != nil {
			g.Logger.Error("failed to unmarshal cached data",
				slog.Any("error", err),
				slog.String("cacheKey", cacheKey),
			)
		}
		if err == nil {
			return nil, nil
		}
	}
	g.Logger.Debug("not using cache",
		slog.Bool("found", found),
		slog.String("cacheKey", cacheKey),
	)
	r, err := g.NewRequest(ctx, endpoint.Method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := g.NewJSONRequest(ctx, r, jsonData)
	if err != nil {
		return res, err
	}
	_, err = g.writeCache(cacheKey, jsonData)
	if err != nil {
		return res, err
	}

	return res, nil
}
