package newspaper

import (
	"context"
	"fmt"
	"time"
)

func filterValidArticles(ctx context.Context, article Article) (bool, error) {
	return article.Valid, nil
}

// dateRangeString returns the inclusive date range (UTC) used for the newspaper run.
// It is intentionally formatted in ISO-8601 (YYYY-MM-DD) to avoid ambiguity in prompts.
func dateRangeString(ctx context.Context) string {
	endTime := time.Now().UTC()
	startTime := endTime.AddDate(0, 0, -optionsFrom(ctx).DaysBack)

	return fmt.Sprintf("%s to %s (inclusive, UTC)", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
}
