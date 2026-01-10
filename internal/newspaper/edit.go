package newspaper

import (
	"context"
	"log/slog"
	"math/rand"
	"sort"

	"github.com/schraf/assistant/pkg/models"
)

func EditNewspaper(ctx context.Context, articles []Article) (*models.Document, error) {
	doc := models.Document{
		Title: "News Report: " + dateRangeText(optionsFrom(ctx).DaysBack),
	}

	for _, article := range articles {
		title := article.Section + ": " + article.Headline
		doc.AddSection(title, article.Body)
	}

	sort.Slice(doc.Sections, func(i, j int) bool {
		return doc.Sections[i].Title < doc.Sections[j].Title
	})

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

		idx := rand.Intn(len(doc.Sections))
		doc.Sections = append(doc.Sections[:idx], doc.Sections[idx+1:]...)
	}

	slog.Info("editing_finished",
		slog.Int("articles", len(doc.Sections)),
		slog.Int("length", doc.Length()),
		slog.Int("max_length", maxLength),
	)

	return &doc, nil
}
