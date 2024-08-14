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

// TestRandomSingleItem test a single item
func TestRandomEmptyItem(t *testing.T) {

	s := []string{}

	out := RandomItem(s)

	if out != "" {
		t.Error("Not the outcome we want:", out)
	}

	n := []int{}

	out2 := RandomItem(n)

	if out2 != 0 {
		t.Error("Not the outcome we want:", out2)
	}

	i := []any{}

	out3 := RandomItem(i)

	if out3 != nil {
		t.Error("Not the outcome we want:", out3)
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

func TestDateToLayout(t *testing.T) {
	tests := []struct {
		Have string
		Want string
	}{
		{Have: "YYYY-MM-DD", Want: "2006-01-02"},
	}

	for _, test := range tests {
		result := DateToLayout(test.Have)

		if result != test.Want {
			t.Log(test.Have, test.Want)
			t.Errorf("Does not match, %s : %s", test.Have, test.Want)
		}
	}
}
