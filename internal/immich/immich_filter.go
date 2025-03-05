package immich

import (
	"time"

	"github.com/charmbracelet/log"
)

// DateFilter applies date filtering to the search request body based on the configured date filter.
// It sets the TakenAfter and TakenBefore fields of the request body to filter photos
// within the determined date range. If no date filter is configured, it returns without
// modifying the request body.
func DateFilter(requestBody *SearchRandomBody, dateFilter string) {
	if dateFilter == "" {
		return
	}

	dateStart, dateEnd, err := determineDateRange(dateFilter)
	if err != nil {
		log.Error("malformed filter", "err", err)
	} else {
		requestBody.TakenAfter = dateStart.Format(time.RFC3339)
		requestBody.TakenBefore = dateEnd.Format(time.RFC3339)
	}
}
