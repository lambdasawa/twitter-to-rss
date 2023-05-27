package main

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
)

type Feed struct {
	ID string

	*feeds.Feed
}

func NewFeed(id string, payload Payload) *Feed {
	feed := &feeds.Feed{
		Title: fmt.Sprintf("T2F: @%s", id),
		Link: &feeds.Link{
			Href: fmt.Sprintf("https://twitter.com/%s", id),
		},
		Description: fmt.Sprintf("Twitter to Feed: @%s", id),
		Author: &feeds.Author{
			Name: fmt.Sprintf("@%s", id),
		},
		Created: time.Now(),
	}

	feed.Items = []*feeds.Item{
		convertItem(id, payload),
	}

	return &Feed{
		ID:   id,
		Feed: feed,
	}
}

func convertItem(id string, payload Payload) *feeds.Item {
	return &feeds.Item{
		Title: payload.Text,
		Link: &feeds.Link{
			Href: payload.Link,
		},
		Description: payload.Text,
		Author: &feeds.Author{
			Name: fmt.Sprintf("@%s", id),
		},
		Created: time.Now(),
	}
}

const (
	MaxItems = 100
)

func (f *Feed) PrependItem(id string, payload Payload) {
	items := make([]*feeds.Item, 0, len(f.Feed.Items)+1)

	items = append(items, convertItem(f.ID, payload))

	items = append(items, f.Feed.Items...)

	if len(items) > MaxItems {
		items = items[:MaxItems]
	}

	f.Feed.Items = items
}

func (f *Feed) Encode() (string, error) {
	return f.Feed.ToAtom()
}
