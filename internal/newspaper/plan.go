package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
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
		8 to 10 article ideas

		## Section
		Section Title: {{.SectionTitle}}
		Description: {{.SectionDescription}}

		## Task
		1. Use web searches to brainstorm candidate news stores for only this section of the newspaper
		2. List no more than 10 candidate stories to be used for this section
		3. For each candidate story, provide:
			- a working headline
			- a short description of the event
		Present the result in a clear, readable text format. Do not use any HTML, markdown, or JSON.
		`
)

func Plan(ctx context.Context, section Section) (*[]Article, error) {
	endTime := time.Now().UTC()
	startTime := endTime.AddDate(0, 0, -optionsFrom(ctx).DaysBack)
	dateRange := fmt.Sprintf("%s to %s", startTime.Format("Jan 2, 2006"), endTime.Format("Jan 2, 2006"))

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
