package newspaper

import (
	"context"
	"log/slog"
)

const (
	SynthesizeSystemPrompt = `
		You are an expert newspaper journalist. Your task is to take researched
		information about a recent event and synthesize a complete newspaper 
		article. The article should contain at least 5 to 8 paragraphs and each
		paragraph should have at least 4 sentences. Do not include any headings
		or Markdown, HTML, LaTeX, or escape characters. The article should be 
		written in clear, neutral, newspaper-style English.

		The user will provide a Date Range. Treat it as a hard constraint:
		only include events, developments, and data points that occurred within
		the Date Range (inclusive). Do not include background/history outside
		the Date Range. If a claim is undated or the date is ambiguous, omit it.
		`

	SynthesizePrompt = `
		## Date Range (inclusive, UTC)
		{{.DateRange}}

		## Research Notes (source material)
		{{.Research}}

		## Task
		Write the article using ONLY information within the Date Range (inclusive).
		Omit anything outside the range or with unclear timing.
		`
)

func SynthesizeArticle(ctx context.Context, article Article) (*Article, error) {
	prompt, err := BuildPrompt(SynthesizePrompt, PromptArgs{
		"DateRange": dateRangeString(ctx),
		"Research":  article.Research,
	})
	if err != nil {
		slog.Warn("synthesizing_article_prompt_failed",
			slog.String("section", article.Section.Title),
			slog.String("headline", article.Headline),
			slog.String("error", err.Error()),
		)

		article.Valid = false
		return &article, nil
	}

	body, err := ask(ctx, SynthesizeSystemPrompt, *prompt)
	if err != nil {
		slog.Warn("synthesizing_article_failed",
			slog.String("section", article.Section.Title),
			slog.String("headline", article.Headline),
			slog.String("error", err.Error()),
		)

		article.Valid = false
	} else {
		article.Valid = true
		article.Body = *body

		slog.Info("synthesized_article",
			slog.String("section", article.Section.Title),
			slog.String("headline", article.Headline),
			slog.Int("body", len(article.Body)),
		)
	}

	return &article, nil
}
