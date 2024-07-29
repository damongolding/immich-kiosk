package utils

import (
	"net/url"
	"reflect"
	"testing"
)

// TestCombineQueries test to see if referer queries overwrite url queries
func TestCombineQueries(t *testing.T) {
	baseQueries := url.Values{}
	baseQueries.Set("refresh", "60")
	baseQueries.Set("transition", "fade")
	baseQueries.Set("raw", "false")

	refererQueries := "/demo-url?transition=none&image_fit=cover&raw=true"

	q, err := CombineQueries(baseQueries, refererQueries)
	if err != nil {
		t.Error(err)
	}

	//  NOT WORKING
	if !reflect.DeepEqual(q["transition"], []string{"fade", "none"}) || q.Get("image_fit") != "cover" || q.Get("refresh") != "60" || !q.Has("raw") {
		t.Error(q)
	}

}

// TestRandomSingleItem test a single item
func TestRandomSingleItem(t *testing.T) {

	s := []string{"cheese"}

	out := RandomItem(s)

	if out != "cheese" {
		t.Error("Not the outcome we want:", out)
	}

}

// TestRandomStruct get out what we expect
func TestRandomStruct(t *testing.T) {

	type RendomStructDemo struct {
		name string
		age  int
	}

	s := []RendomStructDemo{
		{name: "John", age: 20},
		{name: "Clara", age: 34},
	}

	out := RandomItem(s)

	if reflect.TypeOf(out).String() != "utils.RendomStructDemo" {
		t.Error("Not the outcome we want:", out)
	}

}
