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
	ArticleKnowledgeSystemPrompt = `
		You are an expert news reporter and researcher. Your sole task is
		to search the web and available information to gather facts about
		a specific news story by answering a set of questions.
		`

	ArticleKnowledgePrompt = `
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
		
		## Goal 
		Using web search and reliable sources, gather factual information that
		answers each of the questions above and provides the background needed
		to write a complete, balanced news article about this story. Clearly
		indicate which information answers which question.
		`

	ArticleKnowledgeStructureSystemPrompt = `
		You are an expert news researcher and organizer. Your sole task is
		to take raw gathered notes about a news story and structure them into
		an organized list of (topic, information) pairs.
		`

	ArticleKnowledgeStructurePrompt = `
		## Information Gathered
		{{.Information}}
		`
)

type Knowledge struct {
	Topic       string `json:"topic"`
	Information string `json:"information"`
}

// GenerateArticleKnowledge gathers structured knowledge for a single article.
func GenerateArticleKnowledge(ctx context.Context, assistant models.Assistant, section SectionPlan, article ArticlePlan, opts NewspaperOptions, questions []string) ([]Knowledge, error) {
	slog.InfoContext(ctx, "generating_article_knowledge",
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

	prompt, err := BuildPrompt(ArticleKnowledgePrompt, PromptArgs{
		"SectionType": string(section.Type),
		"DateRange":   dateRange,
		"Location":    opts.Location,
		"Headline":    article.Headline,
		"Summary":     article.Summary,
		"Questions":   questions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed building article knowledge prompt: %w", err)
	}

	response, err := assistant.Ask(ctx, ArticleKnowledgeSystemPrompt, *prompt)
	if err != nil {
		return nil, fmt.Errorf("failed gathering article knowledge: %w", err)
	}

	structuredPrompt, err := BuildPrompt(ArticleKnowledgeStructurePrompt, PromptArgs{
		"Information": *response,
	})
	if err != nil {
		return nil, fmt.Errorf("failed building structured article knowledge prompt: %w", err)
	}

	structuredResponse, err := assistant.StructuredAsk(ctx, ArticleKnowledgeStructureSystemPrompt, *structuredPrompt, KnowledgeSchema())
	if err != nil {
		return nil, fmt.Errorf("failed structuring article knowledge: %w", err)
	}

	var knowledge []Knowledge

	if err := json.Unmarshal(structuredResponse, &knowledge); err != nil {
		return nil, fmt.Errorf("failed parsing structured knowledge: %w", err)
	}

	return knowledge, nil
}

func KnowledgeSchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"topic": map[string]any{
					"type":        "string",
					"description": "A short description of the topic this information covers.",
				},
				"information": map[string]any{
					"type":        "string",
					"description": "A detailed report of information about this topic based on previously researched resources.",
				},
			},
			"required": []string{
				"topic",
				"information",
			},
		},
	}
}
