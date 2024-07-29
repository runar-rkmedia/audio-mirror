package rss

import "encoding/xml"

type Rss struct {
	XMLName    xml.Name `xml:"rss"`
	Text       string   `xml:",chardata"`
	Atom       string   `xml:"atom,attr"`
	Itunes     string   `xml:"itunes,attr"`
	Itunesu    string   `xml:"itunesu,attr"`
	Googleplay string   `xml:"googleplay,attr"`
	Version    string   `xml:"version,attr"`
	Channel    Channel  `xml:"channel"`
}

type Channel struct {
	Text           string     `xml:",chardata"`
	Link           Link       `xml:"link"`
	Language       string     `xml:"language"`
	Copyright      string     `xml:"copyright"`
	WebMaster      string     `xml:"webMaster"`
	ManagingEditor string     `xml:"managingEditor"`
	Image          Image      `xml:"image"`
	Owner          Owner      `xml:"owner"`
	Category       []Category `xml:"category"`
	Keywords       string     `xml:"keywords"`
	Explicit       string     `xml:"explicit"`
	PubDate        string     `xml:"pubDate"`
	Title          string     `xml:"title"`
	Author         string     `xml:"author"`
	Description    string     `xml:"description"`
	Summary        string     `xml:"summary"`
	Subtitle       string     `xml:"subtitle"`
	LastBuildDate  string     `xml:"lastBuildDate"`
	Item           []Item     `xml:"item"`
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
	Text     string      `xml:",chardata"`
	AttrText string      `xml:"text,attr"`
	Category Subcategory `xml:"category"`
}
type Subcategory struct {
	Text     string `xml:",chardata"`
	AttrText string `xml:"text,attr"`
}
type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type Item struct {
	Text        string       `xml:",chardata"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Summary     string       `xml:"summary"`
	Subtitle    string       `xml:"subtitle"`
	Category    ItemCategory `xml:"category"`
	Enclosure   Enclosure    `xml:"enclosure"`
	GUID        string       `xml:"guid"`
	Duration    string       `xml:"duration"`
	PubDate     string       `xml:"pubDate"`
	Link        string       `xml:"link"`
}

type Enclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

type ItemCategory struct {
	Text string `xml:",chardata"`
	Code string `xml:"code,attr"`
}
