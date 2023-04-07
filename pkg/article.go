package pkg

import (
	"net/url"
	"time"
)

// Article represents a news article. The fields title, addr, author and published
// form the original news article. The other fields are used for searching and
// book keeping.
type Article struct {
	ID        string
	Title     string
	Addr      url.URL
	Author    string
	Created   time.Time
	Updated   time.Time
	Published time.Time
	Body      string
	Summary   string
	Tags      []string
	Pers      []string
	Locs      []string
	Orgs      []string
}

// Equal checks if this Article a is equal to the given Article b.
// Only fields from the original article are condsidered. Original properties are
// title, addr, author, date of publication and body. Fields are not been validated.
// All other fields of the article are not been checked.
func (a Article) Equal(b Article) bool {
	if a.Title != b.Title {
		return false
	}
	if a.Addr != b.Addr {
		return false
	}
	if a.Author != b.Author {
		return false
	}
	if a.Published != b.Published {
		return false
	}
	if a.Body != b.Body {
		return false
	}
	return true
}
