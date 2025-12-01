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
	// NewspaperSectionIdeasSystemPrompt is used with Ask() and can leverage web search
	// to brainstorm candidate stories for a single section.
	NewspaperSectionIdeasSystemPrompt = `
		You are an expert newspaper editor and news planner. Your task is to
		use web search and your knowledge of recent events to brainstorm
		candidate news stories for a single newspaper section.
		`

	NewspaperSectionIdeasPrompt = `
		## Date Range
		{{.DateRange}}

		## Length
		{{.Length}}

		## Section
		Section type: {{.SectionType}}
		Description: {{.SectionDescription}}

		## Task
		List significantly more candidate stories than will ultimately be used
		for this section (aim for at least twice the target number). For each
		candidate story, provide:
		- a working headline
		- a short description of the event
		- any key details that will help decide if it should be included.
		Present the result in a clear, readable format (not JSON).`

	// NewspaperPlanSystemPrompt is used with StructuredAsk() to turn the
	// brainstormed ideas into a structured plan for a single section.
	NewspaperPlanSystemPrompt = `
		You are an expert newspaper editor and news planner. Your sole task is
		to read a list of candidate stories for a single section and select the
		best ones to include, then output a structured plan in JSON.
		`

	NewspaperSectionPlanPrompt = `
		## Date Range
		{{.DateRange}}

		## Length
		{{.Length}}

		## Section
		Section type: {{.SectionType}}
		Description: {{.SectionDescription}}

		## Candidate Stories
		{{.Ideas}}

		## Task
		From the candidate stories above, select the specific real news stories
		that should be covered in this section. The number of articles for this
		section must match the requested length:
		- short: 3 articles
		- medium: 5 articles
		- long: 8 articles

		For each selected article, provide:
		- "headline": a concise, informative headline
		- "slug": a short identifier for the story
		- "summary": 1–3 sentences summarizing the story and its angle
		- "questions": a list of detailed research questions that must be
		  answered to write a complete news article about this story.
		`
)

// GenerateNewspaperPlan asks the assistant to produce a structured plan of
// sections and articles for a newspaper edition.
func GenerateNewspaperPlan(ctx context.Context, assistant models.Assistant, opts NewspaperOptions) (*NewspaperPlan, error) {
	slog.InfoContext(ctx, "generating_newspaper_plan")

	length := "short"
	switch opts.Depth {
	case ResearchDepthMedium:
		length = "medium"
	case ResearchDepthLong:
		length = "long"
	}

	sectionsMeta := []struct {
		Type        SectionType
		Title       string
		Description string
	}{
		{SectionLocal, "Local News", "Stories focused on the provided location at a state level."},
		{SectionUS, "US News", "Major news stories from across the United States."},
		{SectionWorld, "World News", "Significant international events and developments."},
		{SectionBusiness, "Business and Financial", "Business, markets, and financial news."},
		{SectionTechnology, "Technology", "Technology industry, innovation, and digital trends."},
		{SectionHealthScience, "Health and Science", "Health, medicine, and scientific discoveries."},
	}

	// Compute a human-readable date range from DaysBack for prompts and plan metadata.
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -opts.DaysBack)

	var dateRange string
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		dateRange = end.Format("Jan 2, 2006")
	} else {
		dateRange = fmt.Sprintf("%s–%s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}

	plan := &NewspaperPlan{
		DateRange: dateRange,
		Location:  opts.Location,
		Sections:  make([]SectionPlan, 0, len(sectionsMeta)),
	}

	for _, meta := range sectionsMeta {
		// Only the Local section should be explicitly tied to the user-provided
		// location. We bake the location into the section description for that
		// section, and keep all other sections location-neutral.
		sectionDescription := meta.Description
		if meta.Type == SectionLocal && opts.Location != "" {
			sectionDescription = fmt.Sprintf("Local news stories focused on %s at a state level.", opts.Location)
		}

		// First, brainstorm candidate stories for this section using Ask().
		ideasPrompt, err := BuildPrompt(NewspaperSectionIdeasPrompt, PromptArgs{
			"DateRange":          dateRange,
			"Length":             length,
			"SectionType":        string(meta.Type),
			"SectionDescription": sectionDescription,
		})
		if err != nil {
			return nil, fmt.Errorf("failed building section ideas prompt for %s: %w", meta.Type, err)
		}

		ideasResponse, err := assistant.Ask(ctx, NewspaperSectionIdeasSystemPrompt, *ideasPrompt)
		if err != nil {
			return nil, fmt.Errorf("failed generating section ideas for %s: %w", meta.Type, err)
		}

		// Then, structure and select the final articles using StructuredAsk().
		prompt, err := BuildPrompt(NewspaperSectionPlanPrompt, PromptArgs{
			"DateRange":          dateRange,
			"Length":             length,
			"SectionType":        string(meta.Type),
			"SectionDescription": sectionDescription,
			"Ideas":              *ideasResponse,
		})
		if err != nil {
			return nil, fmt.Errorf("failed building section plan prompt for %s: %w", meta.Type, err)
		}

		response, err := assistant.StructuredAsk(ctx, NewspaperPlanSystemPrompt, *prompt, NewspaperSectionPlanSchema())
		if err != nil {
			return nil, fmt.Errorf("failed building section plan for %s: %w", meta.Type, err)
		}

		var section SectionPlan
		if err := json.Unmarshal(response, &section); err != nil {
			return nil, fmt.Errorf("failed parsing section plan for %s: %w", meta.Type, err)
		}

		// Ensure the type and title are set or overridden according to our metadata.
		section.Type = meta.Type
		if section.Title == "" {
			section.Title = meta.Title
		}

		plan.Sections = append(plan.Sections, section)
	}

	return plan, nil
}

// NewspaperSectionPlanSchema describes the expected JSON structure for a single section plan.
func NewspaperSectionPlanSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"section_type": map[string]any{
				"type": "string",
				"enum": []string{
					string(SectionLocal),
					string(SectionUS),
					string(SectionWorld),
					string(SectionBusiness),
					string(SectionTechnology),
					string(SectionHealthScience),
				},
				"description": "The type of this section.",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "A human-readable title for the section.",
			},
			"articles": map[string]any{
				"type":        "array",
				"description": "A list of article plans for this section.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"headline": map[string]any{
							"type":        "string",
							"description": "A concise, informative headline for the article.",
						},
						"slug": map[string]any{
							"type":        "string",
							"description": "A short identifier for the story.",
						},
						"summary": map[string]any{
							"type":        "string",
							"description": "1–3 sentence summary of the story and its angle.",
						},
						"questions": map[string]any{
							"type":        "array",
							"description": "A list of research questions for this story.",
							"items": map[string]any{
								"type": "string",
							},
						},
					},
					"required": []string{
						"headline",
						"slug",
						"summary",
						"questions",
					},
				},
			},
		},
		"required": []string{
			"section_type",
			"title",
			"articles",
		},
	}
}
