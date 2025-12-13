package newspaper

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

const (
	SynthesizeSystemPrompt = `
		You are an expert newspaper journalist. Your task is to take researched
		information about a recent event and synthesize a complete newspaper 
		article. The article should contain at least 3 to 5 paragraphs and each
		paragraph should have at least 4 sentences. No not include any heading
		or Markdown, HTML, LaTeX, or escape characters. The article should be 
		written in clear, neutral, newspaper-style English.
		`
)

func (p *Pipeline) SynthesizeArticle(ctx context.Context, in <-chan Article, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for article := range in {
				body, err := p.assistant.Ask(ctx, SynthesizeSystemPrompt, article.Research)
				if err != nil {
					return fmt.Errorf("synthesize article error: assistant ask: %w", err)
				}

				article.Body = *body

				slog.Info("synthesized_article",
					slog.String("section", article.Section),
					slog.String("headline", article.Headline),
					slog.Int("body", len(article.Body)),
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
