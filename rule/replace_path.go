package rule

import (
	"net/http"
	"strings"
)

// ReplacePath is a RequestUpdater which
type ReplacePath struct {
	Name    string `json:"name"`
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ RequestUpdater = (*ReplacePath)(nil)

func (rp *ReplacePath) Director(director Director) Director {
	return func(r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, rp.Search, rp.Replace, rp.Times)
		director(r)
	}
}
