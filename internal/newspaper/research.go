package newspaper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/schraf/assistant/pkg/models"
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

func ResearchArticle(ctx context.Context, article Article) (*Article, error) {
	prompt, err := BuildPrompt(ResearchPrompt, PromptArgs{
		"Section":  article.Section,
		"Headline": article.Headline,
		"Summary":  article.Summary,
	})
	if err != nil {
		return nil, fmt.Errorf("research prompt error: %w", err)
	}

	research, err := ask(ctx, ResearchSystemPrompt, *prompt)
	if err != nil {
		if errors.Is(err, models.ErrContentBlocked) {
			slog.Warn("research_content_blocked",
				slog.String("section", article.Section),
				slog.String("headline", article.Headline),
			)

			article.Valid = false
		} else {
			slog.Warn("research_failed",
				slog.String("section", article.Section),
				slog.String("headline", article.Headline),
				slog.String("error", err.Error()),
			)

			article.Valid = false
		}
	} else {
		if len(*research) == 0 {
			slog.Warn("empty_research",
				slog.String("section", article.Section),
				slog.String("headline", article.Headline),
			)

			article.Valid = false
		} else {
			article.Valid = true
			article.Research = *research

			slog.Info("researched_article",
				slog.String("section", article.Section),
				slog.String("headline", article.Headline),
				slog.Int("research", len(article.Research)),
			)
		}
	}

	return &article, nil
}
