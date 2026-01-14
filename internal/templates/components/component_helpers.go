package components

import (
	"strings"

	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

func CreateDataTag(tags []immich.Tag) string {
	clean := make([]string, 0, len(tags))

	for _, tag := range tags {
		clean = append(clean, utils.SanitizeClassName(tag.Value))
	}

	return strings.Join(clean, " ")

}
