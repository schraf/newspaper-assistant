package newspaper

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

const (
	EditSystemPrompt = `
		You are an expert editor. Your role is to review newspaper articles and 
		ensure that each article takes a neutral stance and is clearly written.
		Remove any Markdown, LaTeX, HTML tags, or any escape characters. The
		article should not include any headings. 
		`

	EditShortenSystemPrompt = `
		You are an expert editor. You role is to reduce the length of a
		newspaper article to ensure it is not over 2500 characters.
		The provided article draft is over that limit.
		`
)

func (p *Pipeline) EditArticle(ctx context.Context, in <-chan Article, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for article := range in {
				body, err := p.assistant.Ask(ctx, EditSystemPrompt, article.Body)
				if err != nil {
					return fmt.Errorf("edit article error: assistant ask: %w", err)
				}

				article.Body = *body
				originalLength := len(article.Body)
				attempt := 0

				for len(article.Body) > 2500 {
					attempt++

					body, err := p.assistant.Ask(ctx, EditShortenSystemPrompt, article.Body)
					if err != nil {
						return fmt.Errorf("edit article error: assistant ask: %w", err)
					}

					if len(*body) < len(article.Body) {
						article.Body = *body
					}

					if len(article.Body) <= 2500 {
						break
					}

					if attempt == 5 {
						break
					}
				}

				slog.Info("edited_article",
					slog.String("section", article.Section),
					slog.String("headline", article.Headline),
					slog.Int("original_length", originalLength),
					slog.Int("final_length", len(article.Body)),
				)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case out <- article:
				}
			}

			return nil
		})
	}

	return group.Wait()
}
