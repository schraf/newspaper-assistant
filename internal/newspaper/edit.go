package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/schraf/assistant/pkg/models"
)

const (
	NewspaperEditSystemPrompt = `
		You are an expert newspaper editor-in-chief. Your role is to review
		newspaper editions and ensure they read as cohesive, well-structured
		publications with consistent tone and no unnecessary repetition between
		sections.
		`

	NewspaperEditPrompt = `
		# Newspaper Edition to Edit

		## Sections

		{{range $index, $section := .Sections}}
		### Section {{$index}}: {{$section.Title}}

		{{range $_, $paragraph := $section.Paragraphs}}
		{{$paragraph}}

		{{end}}
		{{end}}

		# Goal
		Review and edit this newspaper edition to ensure:
		1. The edition flows cohesively from section to section.
		2. There is no unnecessary repetition of content between sections.
		3. Each section contributes unique value and has a clear purpose.
		4. Transitions between sections are smooth and logical.
		5. The overall tone and style are consistent and appropriate for a
		   newspaper.
		6. Create a title for the newspaper
		7. Make up a author name for the newspaper

		Maintain the same structure (title and sections with paragraphs) but
		refine the content to eliminate redundancy, improve clarity, and enhance
		readability. If information is repeated across sections, consolidate it
		appropriately or remove redundant instances.
		`
)

// EditNewspaper runs a final editing pass over the synthesized newspaper
// document.
func EditNewspaper(ctx context.Context, assistant models.Assistant, doc *models.Document) (*models.Document, error) {
	slog.InfoContext(ctx, "editing_newspaper")

	prompt, err := BuildPrompt(NewspaperEditPrompt, doc)
	if err != nil {
		return nil, fmt.Errorf("failed building newspaper edit prompt: %w", err)
	}

	response, err := assistant.StructuredAsk(ctx, NewspaperEditSystemPrompt, *prompt, EditNewspaperSchema())
	if err != nil {
		return nil, fmt.Errorf("failed editing newspaper: %w", err)
	}

	var editedDoc models.Document

	if err := json.Unmarshal(response, &editedDoc); err != nil {
		return nil, fmt.Errorf("failed parsing edited newspaper: %w", err)
	}

	return &editedDoc, nil
}

func EditNewspaperSchema() map[string]any {
	// Preserve the original full-document schema used before synthesis was
	// changed to operate on a per-section basis. Editing still works on the
	// complete newspaper document (title + sections).
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "The title of the newspaper edition.",
			},
			"author": map[string]any{
				"type":        "string",
				"description": "The editors name.",
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
