package newspaper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"

	"github.com/schraf/assistant/pkg/models"
)

const (
	EditSystemPrompt = `
		You are an expert newspaper editor. Your task is to trim down the newspaper
		to fit within a specific length, while retaining the most important articles.
		`

	EditPrompt = `
		## Max Length
		{{.MaxLength}}

		## Current Length
		{{.CurrentLength}}

		## Articles
		{{.Articles}}

		## Task
		Review the list of articles and their lengths. Decide which single article
		to remove to help bring the total length closer to the maximum, while
		sacrificing the least amount of important content. The list of articles is 
		provided in a markdown table format.
		`
)

func EditNewspaper(ctx context.Context, articles []Article) (*models.Document, error) {
	doc := models.Document{}

	for _, article := range articles {
		doc.AddSection(article.Headline, article.Body)
	}

	maxLength := optionsFrom(ctx).MaxLength

	slog.Info("editing_start",
		slog.Int("articles", len(articles)),
		slog.Int("length", doc.Length()),
		slog.Int("max_length", maxLength),
	)

	for doc.Length() > maxLength {
		if len(doc.Sections) == 0 {
			break
		}

		var articlesTable strings.Builder
		articlesTable.WriteString("| Index | Headline | Length |\n")
		articlesTable.WriteString("|---|---|---|\n")

		for index, section := range doc.Sections {
			length := 0

			for _, paragraph := range section.Paragraphs {
				length += len(paragraph)
			}

			articlesTable.WriteString(fmt.Sprintf("| %d | %s | %d |\n", index, section.Title, length))
		}

		prompt, err := BuildPrompt(EditPrompt, PromptArgs{
			"MaxLength":     maxLength,
			"CurrentLength": doc.Length(),
			"Articles":      articlesTable.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("edit newspaper prompt error: %w", err)
		}

		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{
					"type":        "integer",
					"description": "The index of the article to remove from the list.",
				},
			},
			"required": []string{"index"},
		}

		var sectionToRemove struct{ Index int }

		responseJson, err := structuredAsk(ctx, EditSystemPrompt, *prompt, schema)
		if err != nil {
			slog.Warn("edit_ask_failed",
				slog.String("error", err.Error()),
			)

			sectionToRemove.Index = rand.Intn(len(doc.Sections))
		} else {
			if err := json.Unmarshal(responseJson, &sectionToRemove); err != nil {
				slog.Warn("edit_ask_unmarshal_failed",
					slog.String("error", err.Error()),
				)

				sectionToRemove.Index = rand.Intn(len(doc.Sections))
			} else {
				if sectionToRemove.Index < 0 || sectionToRemove.Index >= len(doc.Sections) {
					slog.Warn("edit_ask_index_invalid",
						slog.Int("index", sectionToRemove.Index),
					)

					sectionToRemove.Index = rand.Intn(len(doc.Sections))
				}
			}
		}

		removedArticleTitle := doc.Sections[sectionToRemove.Index].Title
		doc.Sections = append(doc.Sections[:sectionToRemove.Index], doc.Sections[sectionToRemove.Index+1:]...)

		slog.Info("removed article",
			slog.String("removed_article_title", removedArticleTitle),
			slog.Int("remaining_articles", len(doc.Sections)),
			slog.Int("length", doc.Length()),
			slog.Int("max_length", maxLength),
		)
	}

	slog.Info("editing_finished",
		slog.Int("articles", len(doc.Sections)),
		slog.Int("length", doc.Length()),
		slog.Int("max_length", maxLength),
	)

	return &doc, nil
}
