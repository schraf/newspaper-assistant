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
		`

	EditPrompt = `
		# Newspaper Article to Edit

		## Section
		{{.Section}}

		## Article Headline
		{{.Headline}}

		## Draft Article
		{{.Body}}

		# Goal
		Review and edit this newspaper article to ensure:
		1. The text flows cohesively in each paragraph
		2. There is no repetition of content between paragraphs
		3. The overall article reads as a unified, well-structured article
		4. The report is no longer than {{.MaxLength}} characters.
		7. Remove any markdown, LaTeX, HTML tags, or any escape characters.
		`
)

func (p *Pipeline) EditArticle(ctx context.Context, in <-chan Article, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for article := range in {
				success := false

				for i := 0; i < 5; i++ {
					prompt, err := BuildPrompt(EditPrompt, PromptArgs{
						"Section":   article.Section,
						"Headline":  article.Headline,
						"Body":      article.Body,
						"MaxLength": 2500 - (i * 200),
					})
					if err != nil {
						return fmt.Errorf("edit article error: %w", err)
					}

					body, err := p.assistant.Ask(ctx, EditSystemPrompt, *prompt)
					if err != nil {
						return fmt.Errorf("edit article error: assistant ask: %w", err)
					}

					article.Body = *body

					if len(article.Body) <= 2500 {
						slog.Info("edited_article",
							slog.String("section", article.Section),
							slog.String("headline", article.Headline),
							slog.Int("body", len(article.Body)),
						)

						success = true

						break
					}
				}

				if !success {
					return fmt.Errorf("edit article error: failed to reduce article to be below 2500 characters")
				}

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
