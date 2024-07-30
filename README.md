# Audio-mirror

## Notes, useful links, references

- [Audio requirements - Apple Podcasts for Creators](https://podcasters.apple.com/support/893-audio-requirements)
- [Spotify Provider Support](https://providersupport.spotify.com/article/podcast-delivery-specification-1-9)
- [RSS 2.0 Specification (RSS 2.0 at Harvard Law)](https://cyber.harvard.edu/rss/rss.html) 
- [Open Podcast API | Open Podcast API](https://openpodcastapi.org/) 


## Features

- [X] Retrieve podcasts from proprietary solutions, like Untold.
- [ ] Serve podcasts to any podcast-player.
- [ ] [Podcast-RSS-compliant "API"](#rss-api)

## Nomenclature

Audio-mirror adheres(TODO) to the RSS-specification for podcasts, and tries to
reuse this nomenclature for all items.

Since Audio-mirror is more general than for podcast, e.g. supports audio-books
as well, the API needs to be a bit more general.

| RSS | API |Podcasts | Audio Books | Description
| ------------- | -------------- | -------------- | ----- | -- |
| channel | channel | Podcast | Book | The general collection |
| ? | episodes | Episode | Part* | The items within a collection. These are often the mediafiles with the accompanying metadata. *Not all books are split in this manner. |
| ? | chapter | chapter | Chapter* | Sections with a media-file. *Not all books are split in this manner |

## RSS-API

> This section is not yet implemented, and mainly focuses on ideas

RSS-feeds are created on demand, which means one can create one by hand.

```
# Serves a podcast with this unique id
/rss/pod/50f371ee-181b-42cd-b2d5-2aa334993640 
# Serves the best match for this podcast, given this search
/rss/pod/GoTime 
# Books, title search with search for author. 
/rss/books/Dexter?author=Lindsay
```
