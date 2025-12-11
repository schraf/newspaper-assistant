package newspaper

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	paragraphPattern *regexp.Regexp
)

func init() {
	paragraphPattern = regexp.MustCompile(`\r?\n\s*\r?\n`)
}

func splitParagraphs(in string) []string {
	paragraphs := paragraphPattern.Split(in, -1)

	for index, paragraph := range paragraphs {
		paragraphs[index] = strings.TrimSpace(paragraph)
	}

	return paragraphs
}

func dateRangeText(daysBack int) string {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -daysBack)

	var dateRange string
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		dateRange = end.Format("Jan 2, 2006")
	} else {
		dateRange = fmt.Sprintf("%s to %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}

	return dateRange
}
