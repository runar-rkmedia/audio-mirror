package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
)

type Rss struct {
	XMLName    xml.Name `xml:"rss"`
	Atom       string   `xml:"xmlns:atom,attr"`
	Itunes     string   `xml:"xmlns:itunes,attr"`
	Itunesu    string   `xml:"xmlns:itunesu,attr"`
	Podcast    string   `xml:"xmlns:podcast,attr"`
	Googleplay string   `xml:"xmlns:googleplay,attr"`
	Version    string   `xml:"version,attr"`
	Channel    Channel  `xml:"channel"`
}

type Channel struct {
	// Required fields

	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        Link   `xml:"link"`
	Language    string `xml:"language"`
	// Strict requirement for values. https://podcasters.apple.com/support/1691-apple-podcasts-categories
	// TODO: add a method that ensures correct categories
	// max 2
	Category []Category `xml:"itunes:category"`
	Explicit string     `xml:"explicit"`
	Image    Image      `xml:"image"`

	// Recommended fields

	// Tells podcast hosting platforms whether they are allowed to import this feed
	Locked string `xml:"podcast:locked"`
	// The globally unique identifier (GUID) for a podcast. The value is a UUIDv5, and generated from the RSS feed URL, with the protocol scheme and trailing slashes stripped off, combined with a unique "podcast" namespace which has a UUID of ead4c236-bf58-58c6-a2c6-a6b28d128cb6.
	GUID string `xml:"podcast:guid"`
	// The group, person, or people responsible for creating the podcast.
	Author string `xml:"itunes:author"`

	// Optional fields

	// The copyright details for a podcast.
	Copyright string `xml:"copyright"`
	// A free-form text field to present a string in a podcast fee
	PodcastTxt PodastTXT `xml:"podcast:txt"`
	// This element specifies the donation/funding links for the podcast. The content of the tag is the recommended string to be used with the link.
	PodcastFunding PodcastFunding `xml:"podcast:funding"`
	// Specifies the podcast as either episodic or se
	// episodic is the default and assumed if this element is not present. This element is required for serial podcasts.
	Type string `xml:"itunes:episodic"`
	// Specifies that a podcast is complete and will not post any more episodes in the future.
	// The only valid value for this element is yes. All other values will be ignored.
	Complete string `xml:"itunes:complete"`

	WebMaster      string `xml:"webMaster"`
	ManagingEditor string `xml:"managingEditor"`
	Owner          Owner  `xml:"owner"`
	Keywords       string `xml:"keywords"`
	PubDate        string `xml:"pubDate"`
	Summary        string `xml:"summary"`
	Subtitle       string `xml:"subtitle"`
	LastBuildDate  string `xml:"lastBuildDate"`
	Item           []Item `xml:"item"`
}

type PodastTXT struct {
	Text    string `xml:",chardata"`
	Purpose string `xml:"purpose,attr"`
}
type PodcastFunding struct {
	Text string `xml:",chardata"`
	URL  string `xml:"url,attr"`
}

func (c Channel) Validate() (err error) {
	err = errors.Join(
		req("Title", c.Title),
		req("Description", c.Description),
		req("Link.Href", c.Link.Href),
		req("Language", c.Language),
		minmax("Category", len(c.Category), 1, 2),
		req("Explicit", c.Explicit),
		req("Image", c.Image.URL),
		oneOf("Type", c.Type, "", "episodic", "serial"),
		oneOf("Complete", c.Complete, "", "yes"),
	)
	return err
}

func req[T comparable](fieldname string, value T) error {
	if value == *new(T) {
		return ErrReq(fieldname)
	}
	return nil
}

func oneOf[T comparable](fieldname string, value T, options ...T) error {
	if slices.Contains(options, value) {
		return nil
	}
	return fmt.Errorf("%s must be one of %v", fieldname, options)
}

func minmax(fieldname string, length, min, max int) error {
	if length < min {
		return fmt.Errorf("%s has invalid length, must be greater than %d, was %d", fieldname, min, length)
	}
	if length > max {
		return fmt.Errorf("%s has invalid length, must be less than or equal to %d, was %d", fieldname, max, length)
	}
	return nil
}

func ErrReq(fieldName string) error {
	return fmt.Errorf("%s is required", fieldName)
}

func RssHeader(channel Channel) Rss {
	// I have not checked what any of these fields actually mean
	return Rss{
		XMLName:    xml.Name{Space: "space", Local: "local"},
		Atom:       "http://www.w3.org/2005/Atom",
		Itunes:     "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Itunesu:    "http://www.itunesu.com/feed",
		Podcast:    "https://podcastindex.org/namespace/1.0",
		Googleplay: "http://www.google.com/schemas/play-podcasts/1.0",
		Version:    "2.0",
		Channel:    channel,
	}
}

type Image struct {
	Text  string `xml:",chardata"`
	Href  string `xml:"href,attr"`
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}
type Owner struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name"`
	Email string `xml:"email"`
}
type Category struct {
	AttrText string      `xml:"text,attr"`
	Category Subcategory `xml:"category"`
}
type Subcategory struct {
	AttrText string `xml:"text,attr"`
}
type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type Item struct {
	Title             string       `xml:"title"`
	Description       string       `xml:"description"`
	Summary           string       `xml:"summary"`
	Subtitle          string       `xml:"subtitle"`
	Category          ItemCategory `xml:"category"`
	Enclosure         Enclosure    `xml:"enclosure"`
	GUID              string       `xml:"guid"`
	DurationInSeconds string       `xml:"duration"`
	PubDate           string       `xml:"pubDate"`
	Link              string       `xml:"link"`
	// Not par
	Image Image `xml:"-"`
}

type Enclosure struct {
	Text          string `xml:",chardata"`
	URL           string `xml:"url,attr"`
	Type          string `xml:"type,attr"`
	LengthInBytes string `xml:"length,attr"`
}

type ItemCategory struct {
	Text string `xml:",chardata"`
	Code string `xml:"code,attr"`
}
