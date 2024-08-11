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

	"github.com/Masterminds/sprig/v3"
	"github.com/runar-rkmedia/audio-mirror/rss"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

	HttpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	GenAPIEndpoint struct {
		Path        string
		Query       string
		Method      string
		RootMapping string
		Mapping     map[string]string
	}
	GenAPI struct {
		Name      string
		CacheTime time.Duration
		Endpoint
		GenAPIOptions
		EndpointSearchTitles *GenAPIEndpoint
		EndpointCategories   *GenAPIEndpoint
		EndpointListEpisodes *GenAPIEndpoint
	}
	GenAPIOptions struct {
		Logger *slog.Logger
		Client HttpClient
		Cache  Cache
	}
	GenAPIChannelList struct {
		Channels []GenApiChannel
		Raw      []byte
	}
	GenApiChannel struct {
		rss.Channel
		Meta GenApiChannelMeta `json:"_meta"`
	}
	GenApiChannelMeta struct {
		Kind      ChannelType
		LastAired *time.Time
		Frequency string
		Source    string
		SourceURL string
		// Id at the source-api. This is typically not the podcast-id, nor audio-mirror's id
		ID string
	}
	ChannelType string
)

var (
	ChannelTypePodCast   ChannelType = "podcast"
	ChannelTypeAudioBook ChannelType = "book"
)

func (g GenAPIEndpoint) CompositeKey() string {
	s := ""
	if g.Method != "" {
		s += g.Method + " "
	}
	u := url.URL{Path: g.Path, RawQuery: g.Query}
	return s + u.String()
}

func init() {
	gjson.AddModifier("literal", func(jsonStr, arg string) string {
		return arg
	})
	gjson.AddModifier("categories", func(jsonStr, arg string) string {
		var entries []string
		err := json.Unmarshal([]byte(jsonStr), &entries)
		if err != nil {
			return err.Error()
		}

		cats := make([]rss.Category, len(entries))
		for i, str := range entries {
			cats[i] = rss.Category{AttrText: str}
		}

		out, err := json.Marshal(cats)
		if err != nil {
			return err.Error()
		}
		return string(out)
	})
}

type GenAPIEpisodeList struct {
	Items []rss.Item
	Raw   []byte
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
		CacheTime:     time.Hour * 1440,
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

func (g *GenAPI) NewJSONRequest(ctx context.Context, r *http.Request, v any, cacheKey string) (*http.Response, error) {
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
	if cacheKey != "" {
		_, err = g.writeCache(cacheKey, body)
		if err != nil {
			return res, err
		}
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		l.Debug("Raw body", slog.String("body", string(body)), slog.Any("error", err))
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	return res, nil
}

func (g *GenAPI) ListTitles(ctx context.Context) (GenAPIChannelList, error) {
	return GenAPIChannelList{}, fmt.Errorf("%w: ListTitles", ErrNotImplemented)
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
	tmpl, err := template.New("_").Funcs(sprig.FuncMap()).Parse(templateString)
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
	fmt.Println("cache?", g.CacheTime, time.Now().Add(-1*g.CacheTime), g.Cache)
	data, ok, err := g.Cache.Retrieve(keyPaths, time.Now().Add(-1*g.CacheTime))
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
	list := &GenAPIEpisodeList{}

	r, body, err := g.RunEndpoint(ctx, *endpoint, data, "episodes-", &list.Items)
	list.Raw = body
	return list, r, err
}

func FlattenAndDeduplicate(logger *slog.Logger, lists []GenAPIChannelList) (j GenAPIChannelList) {
	if logger == nil {
		logger = slog.Default()
	}
	ids := map[string]int{}
	for _, list := range lists {
		for _, channel := range list.Channels {
			if channel.Meta.ID != "" {
				if index, ok := ids[channel.Meta.ID]; ok {
					existing := j.Channels[index]
					logger.Debug("detected duplicate channel, not adding it", slog.String("id", channel.Meta.ID), slog.String("title", channel.Title), slog.String("existingTitle", existing.Title))
					continue

				}
				ids[channel.Meta.ID] = len(j.Channels)
			}
			j.Channels = append(j.Channels, channel)

		}
	}
	return
}

// This is typically used when the api does not have functionality for listing every podcast in a single request.
// Any such implementation endedding GenApi should in those cases override this method.
func (u *GenAPI) FindAllChannels(ctx context.Context) ([]GenAPIChannelList, error) {
	channelList, err := u.ListTitles(ctx)
	if err != nil {
		return nil, err
	}
	return []GenAPIChannelList{channelList}, nil
}

func (g *GenAPI) SearchTitles(ctx context.Context, query string) ([]struct{ Name, ID string }, *http.Response, error) {
	panic("not implemented")
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
	r, _, err := g.RunEndpoint(ctx, *endpoint, data, "search-", &j)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run endpoint: %w", err)
	}
	return j, r, err
}

func (g *GenAPI) DecodeEndpointData(ctx context.Context, endpoint GenAPIEndpoint, dateType string, data []byte, out any) error {
	// var v any
	switch dateType {
	case "", "json":

	default:
		return fmt.Errorf("unknown format: '%s' for deserialization", dateType)
	}
	root := endpoint.RootMapping
	if root == "" {
		root = "@this"
	}
	result := gjson.GetBytes(data, root)
	if !result.IsArray() {
		return fmt.Errorf("expected result to be an array, but was %s from RootMapping %s", result.Type, root)
	}
	outjson := "[]"
	i := 0
	arr := result.Array()
	for _, value := range arr {

		thisJSON := "{}"
		if resy, err := sjson.Set(thisJSON, "_Meta.source", g.Name); err == nil {
			thisJSON = resy
		} else {
			return err
		}
		endpointCompositeKey := endpoint.CompositeKey()
		if endpointCompositeKey != "" {
			if resy, err := sjson.Set(thisJSON, "_Meta.sourceUrl", endpointCompositeKey); err == nil {
				thisJSON = resy
			} else {
				return err
			}
		}
		for key, path := range endpoint.Mapping {
			var v any
			// Escape-hatches
			if strings.HasPrefix(path, "@@template ") {
				out, err := g.TemplateString(strings.TrimPrefix(path, "@template "), map[string]any{"data": data})
				if err != nil {
					return err
				}
				v = out
			} else {
				// treat as gjson-path
				r := value.Get(path)
				switch r.Type {
				case gjson.String:
					v = r.String()
				case gjson.JSON:
					err := json.Unmarshal([]byte(r.Raw), &v)
					if err != nil {
						return err
					}
				case gjson.Null:
					continue
				default:
					return fmt.Errorf("unhandled type %s", r.Type)
				}
			}
			resy, err := sjson.Set(thisJSON, key, v)
			if err != nil {
				return err
			}
			thisJSON = resy

		}
		if thisJSON != "{}" {
			var j any
			err := json.Unmarshal([]byte(thisJSON), &j)
			if err != nil {
				return err
			}

			updated, err := sjson.Set(outjson, "-1", j)
			if err != nil {
				return err
			}
			outjson = updated
			i++
		}
	}
	err := json.Unmarshal([]byte(outjson), out)
	if err != nil {
		return err
	}
	// var raw any
	// json.Unmarshal([]byte(outjson), &raw)
	// pretty, _ := json.MarshalIndent(raw, "", "  ")
	// fmt.Println("outjson", string(pretty))
	return nil
}

func (g *GenAPI) RunEndpoint(
	ctx context.Context,
	endpoint GenAPIEndpoint,
	data map[string]any,
	cacheKeyPrefix string,
	responseData any,
) (*http.Response, []byte, error) {
	cacheKey := cacheKeyPrefix
	if data == nil {
		data = map[string]any{}
	}
	for k, v := range data {
		cacheKey += k + "=" + fmt.Sprintf("%v", v)
	}
	url, err := g.CreateSubURL(g.URL, endpoint, data)
	if err != nil {
		return nil, nil, err
	}
	cached, found := g.getCache(cacheKey)
	if found && len(cached) > 0 {
		g.Logger.Debug("using cache",
			slog.String("cacheKey", cacheKey),
		)
		if err == nil {
			// TODO: add the datatype
			err = g.DecodeEndpointData(ctx, endpoint, "", cached, responseData)
			if err == nil {
				return nil, cached, nil
			}
		}
	}
	panic("TEMP DISABLE RUNNER")
	g.Logger.Debug("not using cache",
		slog.Bool("found", found),
		slog.String("cacheKey", cacheKey),
	)
	r, err := g.NewRequest(ctx, endpoint.Method, url.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	l := g.Logger.With(
		slog.String("uri", r.URL.String()),
		slog.String("method", r.Method),
	)
	l.Debug("Performing request")
	res, err := g.Client.Do(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to do request: %w", err)
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
		return nil, nil, fmt.Errorf("failed to read body: %w", err)
	}
	if l.Enabled(ctx, -10) {
		l.Debug("Raw body", slog.String("body", string(body)), slog.Any("headers", res.Header))
	}
	if res.StatusCode >= 400 {
		l.Error("unsuccessful statuscode", slog.String("body", string(body)))
		return nil, nil, fmt.Errorf("unsuccessful status-code: %d", res.StatusCode)
	}
	if cacheKey != "" {
		_, err = g.writeCache(cacheKey, body)
		if err != nil {
			return res, body, err
		}
	}

	err = g.DecodeEndpointData(ctx, endpoint, "", body, responseData)
	if err != nil {
		return res, body, err
	}
	return res, body, err
}
