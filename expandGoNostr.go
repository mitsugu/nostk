package main

import (
	"github.com/nbd-wtf/go-nostr"
)

// expand nostr.Tag structure
type ExTag struct {
	tag nostr.Tag
}

func (t *ExTag) addTagName(tn string) {
	t.tag = append(t.tag, tn)
}

func (t *ExTag) addTagValue(v string) {
	t.tag = append(t.tag, v)
}

func (t *ExTag) getNostrTag() nostr.Tag {
	return t.tag
}

// expand nostr.Tags structure
type ExTags struct {
	tags nostr.Tags
}

func (ts *ExTags) addTag(t nostr.Tag) {
	ts.tags = append(ts.tags, t)
}

func (t *ExTags) getNostrTags() nostr.Tags {
	return t.tags
}
