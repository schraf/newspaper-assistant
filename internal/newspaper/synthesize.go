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
		article.
		`

	SynthesizePrompt = `
		## Newspaper Section
		{{.Section}}

		## Article Headline
		{{.Headline}}

		## Event Summary
		{{.Summary}}

		## Researched Information
		{{.Research}}

		## Goal
		Create a newspaper article for the given news event using the
		provided researched information. The article should:
		- Contain at least 3 to 5 paragraphs.
		- Each paragraph should have at least 4 sentences.
		- Be written in clear, neutral, newspaper-style English.
		- The output should not contain any special characters, Markdown, HTML, LaTeX, or escape characters.
		`
)

func (p *Pipeline) SynthesizeArticle(ctx context.Context, in <-chan Article, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for article := range in {
				prompt, err := BuildPrompt(SynthesizePrompt, PromptArgs{
					"Section":  article.Section,
					"Headline": article.Headline,
					"Summary":  article.Summary,
					"Research": article.Research,
				})
				if err != nil {
					return fmt.Errorf("synthesize article error: %w", err)
				}

				body, err := p.assistant.Ask(ctx, SynthesizeSystemPrompt, *prompt)
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
