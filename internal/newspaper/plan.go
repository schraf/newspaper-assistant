package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

const (
	SectionPlanSystemPrompt = `
		You are an expert newspaper editor and news planner. Your task is to
		evaluate news stories for a single newspaper section.
		`

	SectionPlanPrompt = `
		## Date Range
		{{.DateRange}}

		## Length
		3 to 4 article ideas

		## Section
		Section Title: {{.SectionTitle}}
		Description: {{.SectionDescription}}

		## Task
		1. Use web searches to brainstorm candidate news stores for only this section of the newspaper
		2. List no more than 4 candidate stories to be used for this section
		3. For each candidate story, provide:
			- a working headline
			- a short description of the event
		Present the result in a clear, readable text format. Do not use any HTML, markdown, or JSON.
		`
)

func (p *Pipeline) PlanSection(ctx context.Context, in <-chan Section, out chan<- Article, concurrency int) error {
	defer close(out)

	group, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for section := range in {
				dateRange := dateRangeText(p.options.DaysBack)

				prompt, err := BuildPrompt(SectionPlanPrompt, PromptArgs{
					"DateRange":          dateRange,
					"SectionTitle":       section.Title,
					"SectionDescription": section.Description,
				})
				if err != nil {
					return fmt.Errorf("generate section plan error (%s): %w", section.Title, err)
				}

				response, err := p.assistant.Ask(ctx, SectionPlanSystemPrompt, *prompt)
				if err != nil {
					return fmt.Errorf("generate section plan error: assistant ask (%s): %w", section.Title, err)
				}

				schema := map[string]any{
					"type":        "array",
					"description": "list of articles",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"headline": map[string]any{
								"type":        "string",
								"description": "headline of the article",
							},
							"summary": map[string]any{
								"type":        "string",
								"description": "summary of the article",
							},
						},
						"required": []string{"headline", "summary"},
					},
				}

				structuredPrompt := "Extract the list of articles from the following text.\n" + *response

				responseJson, err := p.assistant.StructuredAsk(ctx, SectionPlanSystemPrompt, structuredPrompt, schema)
				if err != nil {
					return fmt.Errorf("generate section plan error: assistant structured ask (%s): %w", section.Title, err)
				}

				var articles []Article

				if err := json.Unmarshal(responseJson, &articles); err != nil {
					return fmt.Errorf("generate section plan error: unmarshal json (%s): %w", section.Title, err)
				}

				if len(articles) > 4 {
					articles = articles[:5]
				}

				headlines := []string{}

				for _, article := range articles {
					headlines = append(headlines, article.Headline)
				}

				slog.InfoContext(ctx, "generated_section_articles",
					slog.String("section", section.Title),
					slog.Int("articles", len(articles)),
					slog.Any("headlines", headlines),
				)

				for _, article := range articles {
					article.Section = section.Title

					select {
					case <-ctx.Done():
						return ctx.Err()
					case out <- article:
					}
				}
			}

			return nil
		})
	}

	return group.Wait()
}
