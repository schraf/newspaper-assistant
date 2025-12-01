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
	AnalyzeArticleKnowledgeSystemPrompt = `
		You are an expert news editor and analyst. Your sole task is to review
		the information gathered for a news story and determine whether it is
		sufficient to write a complete, accurate, and balanced article.
		`

	AnalyzeArticleKnowledgePrompt = `
		## Article Goal
		{{.ArticleGoal}}

		## Section
		Type: {{.SectionType}}

		## Date Range
		{{.DateRange}}

		## Location
		{{.Location}}

		## Article Headline
		{{.Headline}}

		## Article Summary
		{{.Summary}}

		## Questions To Answer
		{{range $index, $question := .Questions}}
		{{$index}}. {{$question}}
		{{end}}

		## Information Gathered
		{{range $index, $knowledge := .Knowledge}}
		{{$index}}. {{$knowledge.Topic}}
		{{$knowledge.Information}}

		{{end}}
		
		## Task
		Review the gathered information and determine whether each question has
		been answered with enough detail and context to support a strong news
		article. If there are gaps, ambiguities, missing perspectives, or
		important facts that are not covered, provide a list of follow-up
		research questions that should be answered for this story. If no
		additional research is needed, return an empty list.
		`
)

// AnalyzeArticleKnowledge inspects gathered knowledge for an article and
// returns follow-up research questions, if any.
func AnalyzeArticleKnowledge(ctx context.Context, assistant models.Assistant, goal string, section SectionPlan, article ArticlePlan, questions []string, knowledge []Knowledge, opts NewspaperOptions) ([]string, error) {
	slog.InfoContext(ctx, "analyzing_article_knowledge",
		slog.String("section", string(section.Type)),
		slog.String("headline", article.Headline),
	)

	end := time.Now().UTC()
	start := end.AddDate(0, 0, -opts.DaysBack)

	var dateRange string
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		dateRange = end.Format("Jan 2, 2006")
	} else {
		dateRange = fmt.Sprintf("%s–%s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}

	prompt, err := BuildPrompt(AnalyzeArticleKnowledgePrompt, PromptArgs{
		"ArticleGoal": goal,
		"SectionType": string(section.Type),
		"DateRange":   dateRange,
		"Location":    opts.Location,
		"Headline":    article.Headline,
		"Summary":     article.Summary,
		"Questions":   questions,
		"Knowledge":   knowledge,
	})
	if err != nil {
		return nil, fmt.Errorf("failed building analyze article knowledge prompt: %w", err)
	}

	response, err := assistant.StructuredAsk(ctx, AnalyzeArticleKnowledgeSystemPrompt, *prompt, AnalyzeKnowledgeSchema())
	if err != nil {
		return nil, fmt.Errorf("failed analyze article knowledge request: %w", err)
	}

	var furtherQuestions []string

	if err := json.Unmarshal(response, &furtherQuestions); err != nil {
		return nil, fmt.Errorf("failed parsing analyze article knowledge response: %w", err)
	}

	return furtherQuestions, nil
}

func AnalyzeKnowledgeSchema() map[string]any {
	return map[string]any{
		"type":        "array",
		"description": "A list of follow up research questions to be answered",
		"items": map[string]any{
			"type":        "string",
			"description": "Research question",
		},
	}
}
