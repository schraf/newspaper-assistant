package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
	SectionPlanSystemPrompt = `
		You are an expert newspaper editor and news planner. Your task is to
		evaluate news stories for a single newspaper section.

		The user will provide a Date Range. Treat it as a hard constraint:
		only propose stories whose primary event/development occurred within
		the Date Range (inclusive). If you cannot verify the timing is within
		range, do not include the story.
		`

	SectionPlanPrompt = `
		## Date Range
		{{.DateRange}}

		IMPORTANT: The Date Range is a hard constraint (inclusive). Only consider events, developments, and data points that occurred within the Date Range. If you cannot clearly verify that an event happened within the Date Range, do not include it.

		## Length
		8 to 10 article ideas

		## Section
		Section Title: {{.SectionTitle}}
		Description: {{.SectionDescription}}

		## Task
		1. Use web searches to brainstorm candidate news stories for only this section of the newspaper
		2. Only propose stories where the primary event/development occurred within the Date Range (inclusive)
		3. If a story spans a longer timeline, only include it if there was a significant, date-verifiable development within the Date Range; otherwise exclude it
		4. Avoid background/history outside the Date Range; do not select anniversary pieces, retrospectives, or "in previous years" recaps
		5. List no more than 10 candidate stories to be used for this section
		6. For each candidate story, provide:
			- a working headline
			- a short description of the event (include the specific in-range date or in-range time window in the description)
		Present the result in a clear, readable text format. Do not use any HTML, markdown, or JSON.
		`
)

func Plan(ctx context.Context, section Section) (*[]Article, error) {
	dateRange := dateRangeString(ctx)

	prompt, err := BuildPrompt(SectionPlanPrompt, PromptArgs{
		"DateRange":          dateRange,
		"SectionTitle":       section.Title,
		"SectionDescription": section.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("generate section plan error (%s): %w", section.Title, err)
	}

	response, err := ask(ctx, SectionPlanSystemPrompt, *prompt)
	if err != nil {
		return nil, fmt.Errorf("generate section plan error: assistant ask (%s): %w", section.Title, err)
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

	responseJson, err := structuredAsk(ctx, SectionPlanSystemPrompt, structuredPrompt, schema)
	if err != nil {
		return nil, fmt.Errorf("generate section plan error: assistant structured ask (%s): %w", section.Title, err)
	}

	var articles []Article

	if err := json.Unmarshal(responseJson, &articles); err != nil {
		return nil, fmt.Errorf("generate section plan error: unmarshal json (%s): %w", section.Title, err)
	}

	if len(articles) == 0 {
		return nil, fmt.Errorf("generate section plan error: no articles generated for section %s", section.Title)
	}

	for index := 0; index < len(articles); index++ {
		articles[index].Valid = true
		articles[index].Section = section

		slog.Info("generated_section_article",
			slog.String("section", section.Title),
			slog.Any("headline", articles[index].Headline),
		)
	}

	return &articles, nil
}
