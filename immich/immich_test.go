package immich

import (
	"testing"

	"github.com/damongolding/immich-kiosk/config"
)

// TestGetRandomImage testing incorrect url
func TestGetRandomImage(t *testing.T) {
	c := config.Config{
		ImmichUrl:    "https://nope.com",
		ImmichApiKey: "123456",
	}

	i := NewImage(c)

	err := i.GetRandomImage("TESTING")
	if err != nil {
		t.Error(err)
	}
}
