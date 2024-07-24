package utils

import (
	"net/url"
	"testing"
)

// TestCombineQueries test to see if referer queries overwrite url queries
func TestCombineQueries(t *testing.T) {
	baseQueries := url.Values{}
	baseQueries.Set("refresh", "60")
	baseQueries.Set("transition", "fade")

	refererQueries := "demo-url?transition=none&fill_screen=true&raw"

	q, err := CombineQueries(baseQueries, refererQueries)
	if err != nil {
		t.Error(err)
	}

	if q.Get("transition") != "none" && q.Get("fill_screen") != "true" && q.Get("refresh") != "60" && !q.Has("raw") {
		t.Error(q)
	}

}
