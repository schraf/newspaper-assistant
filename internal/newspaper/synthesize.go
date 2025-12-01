package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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
		- Contain one section for each individual article listed above (not
		  one per high-level newspaper section).
		- Name each section using the pattern
		  "<Newspaper Section Title> – <Article Headline>" so that every
		  article appears as its own section in the final document.
		- For each section, include a short introductory paragraph followed by
		  the full article body derived from the researched information for
		  that specific article only.
		- Do not merge multiple articles into a single section.
		- Be written in clear, neutral, newspaper-style English.
		Output the edition as JSON matching the provided schema.
		`
)

// SynthesizeNewspaper turns section/article research results into a
// newspaper-style models.Document.
func SynthesizeNewspaper(ctx context.Context, assistant models.Assistant, opts NewspaperOptions, sections []SectionResearch) (*models.Document, error) {
	slog.InfoContext(ctx, "synthesizing_newspaper")

	// We synthesize the final document one high-level section at a time to
	// keep prompts focused and enable more controllable generation. Each
	// per-section synthesis call returns a single section-shaped JSON payload
	// (title + paragraphs), which we later merge into a single newspaper
	// document.

	var doc models.Document

	end := time.Now().UTC()
	start := end.AddDate(0, 0, -opts.DaysBack)

	var dateRange string
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		dateRange = end.Format("Jan 2, 2006")
	} else {
		dateRange = fmt.Sprintf("%s–%s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}

	for i, section := range sections {
		prompt, err := BuildPrompt(NewspaperSynthesizePrompt, PromptArgs{
			"DateRange": dateRange,
			"Location":  opts.Location,
			"Sections":  []SectionResearch{section},
		})
		if err != nil {
			return nil, fmt.Errorf("failed building newspaper synthesis prompt for section %d: %w", i, err)
		}

		response, err := assistant.StructuredAsk(ctx, NewspaperSynthesizeSystemPrompt, *prompt, NewspaperDocumentSectionSchema())
		if err != nil {
			return nil, fmt.Errorf("failed synthesizing newspaper section %d: %w", i, err)
		}

		var docSection models.DocumentSection
		if err := json.Unmarshal(response, &docSection); err != nil {
			return nil, fmt.Errorf("failed parsing synthesized newspaper section %d: %w", i, err)
		}

		doc.Sections = append(doc.Sections, docSection)
	}

	return &doc, nil
}

// NewspaperDocumentSectionSchema describes the JSON structure for a single synthesized
// section. Synthesis now operates on one high-level section at a time using
// this schema, and the individual sections are merged into a full document
// afterwards.
func NewspaperDocumentSectionSchema() map[string]any {
	return map[string]any{
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
	}
}
