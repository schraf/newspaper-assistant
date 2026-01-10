package newspaper

import (
	"context"
)

func filterValidArticles(ctx context.Context, article Article) (bool, error) {
	return article.Valid, nil
}
