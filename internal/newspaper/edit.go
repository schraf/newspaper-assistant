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
		You are an expert newspaper editor. Your role is to review and polish
		individual newspaper articles so they are clear, well-structured,
		readable, and written in a professional newspaper tone while preserving
		the original meaning and factual content.
		`

	NewspaperEditPrompt = `
		# Newspaper Article to Edit

		## Title
		{{.Title}}

		## Paragraphs

		{{range $_, $paragraph := .Paragraphs}}
		{{$paragraph}}

		{{end}}

		# Goal
		Review and edit this single newspaper article to ensure:
		1. The writing is clear, concise, and easy to follow.
		2. The tone and style are appropriate for a professional newspaper.
		3. The structure flows logically from paragraph to paragraph.
		4. Grammar, spelling, and punctuation are corrected.
		5. The original facts and meaning are preserved (do not invent new facts).

		Maintain the same overall structure (a title and a sequence of
		paragraphs) while refining the language for clarity, style, and
		readability.
		`
)

// EditNewspaper runs a final editing pass over each synthesized newspaper
// section independently. This keeps prompts small and focused and avoids
// cross-section context issues.
func EditNewspaper(ctx context.Context, assistant models.Assistant, doc *models.Document) (*models.Document, error) {
	slog.InfoContext(ctx, "editing_newspaper")

	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	// Copy the original document so we can mutate sections in place without
	// touching the caller's instance.
	editedDoc := *doc

	for i, section := range doc.Sections {
		// Build a per-article prompt using just this section's content.
		prompt, err := BuildPrompt(NewspaperEditPrompt, PromptArgs{
			"Title":      section.Title,
			"Paragraphs": section.Paragraphs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed building newspaper edit prompt for section %d: %w", i, err)
		}

		response, err := assistant.StructuredAsk(ctx, NewspaperEditSystemPrompt, *prompt, NewspaperDocumentSectionSchema())
		if err != nil {
			return nil, fmt.Errorf("failed editing newspaper section %d: %w", i, err)
		}

		var editedSection models.DocumentSection

		if err := json.Unmarshal(response, &editedSection); err != nil {
			return nil, fmt.Errorf("failed parsing edited newspaper section %d: %w", i, err)
		}

		editedDoc.Sections[i] = editedSection
	}

	return &editedDoc, nil
}

// EditNewspaperSchema previously described a full-document edit response.
// It is retained for backward compatibility but no longer used now that
// editing operates on individual sections using NewspaperDocumentSectionSchema.
func EditNewspaperSchema() map[string]any {
	return NewspaperDocumentSectionSchema()
}
