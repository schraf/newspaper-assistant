package newspaper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/schraf/assistant/pkg/models"
	"golang.org/x/sync/errgroup"
)

const (
	ResearchSystemPrompt = `
		You are an expert newspaper researcher. Your sole task is
		to search the web to gather information about a provided
		news event.
		`

	ResearchPrompt = `
		## Newspaper Section
		{{.Section}}

		## Article Headline
		{{.Headline}}

		## Event Summary
		{{.Summary}}
		
		## Goal 
		Search the web and gather information about this single
		event given a headline for the article along with the
		event summary. You should provide plenty of information 
		around the event details in a well formatted structure.
		`
)

func (p *Pipeline) ResearchArticle(ctx context.Context, in <-chan Article, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for article := range in {
				prompt, err := BuildPrompt(ResearchPrompt, PromptArgs{
					"Section":  article.Section,
					"Headline": article.Headline,
					"Summary":  article.Summary,
				})
				if err != nil {
					return fmt.Errorf("research article error: %w", err)
				}

				research, err := p.assistant.Ask(ctx, ResearchSystemPrompt, *prompt)
				if err != nil {
					if errors.Is(err, models.ErrContentBlocked) {
						slog.Warn("research_content_blocked",
							slog.String("section", article.Section),
							slog.String("headline", article.Headline),
						)

						continue
					}

					return fmt.Errorf("research article error: assistant ask: %w", err)
				}

				if len(*research) == 0 {
					return fmt.Errorf("research article error: no research found for '%s': %w", article.Headline, err)
				}

				article.Research = *research

				slog.Info("researched_article",
					slog.String("section", article.Section),
					slog.String("headline", article.Headline),
					slog.Int("research", len(article.Research)),
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
