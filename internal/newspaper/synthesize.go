package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/schraf/assistant/pkg/models"
)

const (
	NewspaperSynthesizeSystemPrompt = `
		You are an expert newspaper writer. Your task is to take researched
		information for multiple sections and articles and synthesize a complete
		newspaper edition.
		`

	NewspaperSynthesizePrompt = `
		## Date Range
		{{.DateRange}}

		## Location
		{{.Location}}

		## Researched Sections and Articles

		{{range $_, $section := .Sections}}
		### Section: {{$section.Plan.Title}} ({{$section.Plan.Type}})

		{{range $_, $article := $section.Articles}}
		#### Article: {{$article.Plan.Headline}}
		Planned summary: {{$article.Plan.Summary}}

		{{range $_, $knowledge := $article.Knowledge}}
		- {{$knowledge.Topic}}: {{$knowledge.Information}}
		{{end}}

		{{end}}
		{{end}}

		## Goal 
		Create a structured newspaper edition for the given date range and
		location using the researched information above. The edition should:
		- Have an overall title.
		- Contain one section for each of the listed sections.
		- For each section, include a short introductory paragraph followed by
		  well-written article writeups derived from the researched information.
		- Be written in clear, neutral, newspaper-style English.
		Output the edition as JSON matching the provided schema.
		`
)

// SynthesizeNewspaper turns section/article research results into a
// newspaper-style models.Document.
func SynthesizeNewspaper(ctx context.Context, assistant models.Assistant, opts NewspaperOptions, sections []SectionResearch) (*models.Document, error) {
	slog.InfoContext(ctx, "synthesizing_newspaper")

	prompt, err := BuildPrompt(NewspaperSynthesizePrompt, PromptArgs{
		"DateRange": opts.DateRange,
		"Location":  opts.Location,
		"Sections":  sections,
	})
	if err != nil {
		return nil, fmt.Errorf("failed building newspaper synthesis prompt: %w", err)
	}

	response, err := assistant.StructuredAsk(ctx, NewspaperSynthesizeSystemPrompt, *prompt, NewspaperDocumentSchema())
	if err != nil {
		return nil, fmt.Errorf("failed synthesizing newspaper: %w", err)
	}

	var doc models.Document

	if err := json.Unmarshal(response, &doc); err != nil {
		return nil, fmt.Errorf("failed parsing synthesized newspaper: %w", err)
	}

	return &doc, nil
}

// NewspaperDocumentSchema describes the JSON structure for the final document.
func NewspaperDocumentSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "The title of the newspaper edition.",
			},
			"sections": map[string]any{
				"type":        "array",
				"description": "The sections of the newspaper edition.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "The title of the section.",
						},
						"paragraphs": map[string]any{
							"type":        "array",
							"description": "Paragraphs of text for this section, including intro and article bodies.",
							"items": map[string]any{
								"type": "string",
							},
						},
					},
					"required": []string{
						"title",
						"paragraphs",
					},
				},
			},
		},
		"required": []string{
			"title",
			"sections",
		},
	}
}
