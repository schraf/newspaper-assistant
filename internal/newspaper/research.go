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

		The user will provide a Date Range. Treat it as a hard constraint:
		only gather and report events, developments, and data points that
		occurred within the Date Range (inclusive). If you cannot verify the
		date is within range, omit it.
		`

	ResearchPrompt = `
		## Date Range (inclusive, UTC)
		{{.DateRange}}

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

		## Hard Rules (do not violate)
		- Only include facts/events/data that occurred within the Date Range (inclusive).
		- If sources discuss background/history outside the Date Range, do not include it.
		- If a claim is undated or the date is ambiguous, omit it.
		- Prefer sources that explicitly state dates within the Date Range.

		## Output Requirements
		- Plain text only (no HTML, no Markdown).
		- Include the in-range dates next to key facts/numbers.
		- If you cannot find enough in-range information to support the story, say so explicitly.
		`
)

func ResearchArticle(ctx context.Context, article Article) (*Article, error) {
	prompt, err := BuildPrompt(ResearchPrompt, PromptArgs{
		"DateRange": dateRangeString(ctx),
		"Section":   article.Section.Title,
		"Headline":  article.Headline,
		"Summary":   article.Summary,
	})
	if err != nil {
		return nil, fmt.Errorf("research prompt error: %w", err)
	}

	research, err := ask(ctx, ResearchSystemPrompt, *prompt)
	if err != nil {
		if errors.Is(err, models.ErrContentBlocked) {
			slog.Warn("research_content_blocked",
				slog.String("section", article.Section.Title),
				slog.String("headline", article.Headline),
			)

			article.Valid = false
		} else {
			slog.Warn("research_failed",
				slog.String("section", article.Section.Title),
				slog.String("headline", article.Headline),
				slog.String("error", err.Error()),
			)

			article.Valid = false
		}
	} else {
		if len(*research) == 0 {
			slog.Warn("empty_research",
				slog.String("section", article.Section.Title),
				slog.String("headline", article.Headline),
			)

			article.Valid = false
		} else {
			article.Valid = true
			article.Research = *research

			slog.Info("researched_article",
				slog.String("section", article.Section.Title),
				slog.String("headline", article.Headline),
				slog.Int("research", len(article.Research)),
			)
		}
	}

	return &article, nil
}
