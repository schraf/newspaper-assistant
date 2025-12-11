package newspaper

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/schraf/assistant/pkg/models"
	"golang.org/x/sync/errgroup"
)

type Pipeline struct {
	assistant models.Assistant
	options   NewspaperOptions
}

func NewPipeline(assistant models.Assistant, options NewspaperOptions) *Pipeline {
	return &Pipeline{
		assistant: assistant,
		options:   options,
	}
}

func (p *Pipeline) Exec(ctx context.Context) (*models.Document, error) {
	slog.Info("generating_newspaper",
		slog.Int("days_back", p.options.DaysBack),
		slog.String("location", p.options.Location),
	)

	section := make(chan Section, 6)

	section <- Section{
		Title:       "Local News",
		Description: fmt.Sprintf("Local news stories from %s.", p.options.Location),
	}

	section <- Section{
		Title:       "US News",
		Description: "Major news stories from across the United States.",
	}

	section <- Section{
		Title:       "World News",
		Description: "Significant international events and developments.",
	}

	section <- Section{
		Title:       "Business and Financial",
		Description: "Business, markets, and financial news.",
	}

	section <- Section{
		Title:       "Technology",
		Description: "Technology industry, innovation, and digital trends.",
	}

	section <- Section{
		Title:       "Health and Science",
		Description: "Health, medicine, and scientific discoveries.",
	}

	close(section)

	articles := make(chan Article, 100)
	research := make(chan Article, 100)
	synthesis := make(chan Article, 100)

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return p.PlanSection(ctx, section, articles, 10)
	})

	group.Go(func() error {
		return p.ResearchArticle(ctx, articles, research, 10)
	})

	group.Go(func() error {
		return p.SynthesizeArticle(ctx, research, synthesis, 10)
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	doc := models.Document{
		Title: "News Report: " + dateRangeText(p.options.DaysBack),
	}

	for article := range synthesis {
		section := models.DocumentSection{
			Title:      article.Section + ": " + article.Headline,
			Paragraphs: splitParagraphs(article.Body),
		}

		doc.Sections = append(doc.Sections, section)
	}

	sort.Slice(doc.Sections, func(i, j int) bool {
		return doc.Sections[i].Title < doc.Sections[j].Title
	})

	return &doc, nil
}
