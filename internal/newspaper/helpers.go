package newspaper

import (
	"context"
	"fmt"
	"time"
)

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

func filterValidArticles(ctx context.Context, article Article) (bool, error) {
	return article.Valid, nil
}
