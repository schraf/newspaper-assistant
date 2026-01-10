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
		`
)

func SynthesizeArticle(ctx context.Context, article Article) (*Article, error) {
	body, err := ask(ctx, SynthesizeSystemPrompt, article.Research)
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
