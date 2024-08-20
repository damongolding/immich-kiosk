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
		In   string
		Want string
	}{
		{"YYYY-MM-DD", "2006-01-02"},
		{"YYYY/MM/DD", "2006/01/02"},
		{"YYYY:MM:DD", "2006:01:02"},
		{"YYYY MM DD", "2006 01 02"},
		{"YY M DDD", "06 1 Mon"},
		{"YY MMM DDDD", "06 Jan Monday"},
		{"YYYY MMMM DDDD", "2006 January Monday"},
		{"YYYYYY-MM-DD", "200606-01-02"},
		{"YYYY MM DD additional text", "2006 01 02 additional text"},
		{"", ""},
	}

	for _, test := range tests {
		result := DateToLayout(test.In)

		if result != test.Want {
			t.Log(result, test.Want)
			t.Errorf("Does not match, %s : %s", result, test.Want)
		}
	}
}

func TestStringToColor(t *testing.T) {
	in := "a sample string"

	a := StringToColor(in)
	b := StringToColor(in)

	if a.Hex != b.Hex {
		t.Error("colors do not match")
	}
}

func TestColorContrast(t *testing.T) {
	white := Color{R: 255, G: 255, B: 255}
	black := Color{R: 0, G: 0, B: 0}
	grey := Color{R: 105, G: 105, B: 105}

	maxRatio := CalculateContrastRatio(white, black)
	if maxRatio != 21 {
		t.Error("ratio is not at maximum", maxRatio)
	}

	failRatio := CalculateContrastRatio(grey, black)
	if failRatio == 21 {
		t.Error("ratio is not at maximum", failRatio)
	}

}
