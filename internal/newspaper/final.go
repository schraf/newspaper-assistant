package newspaper

import (
	"context"
	"math/rand"
	"sort"

	"github.com/schraf/assistant/pkg/models"
)

func (p *Pipeline) FinalizeNewspaper(ctx context.Context, in <-chan Article, out chan<- models.Document) error {
	defer close(out)

	doc := models.Document{
		Title: "News Report: " + dateRangeText(p.options.DaysBack),
	}

	for article := range in {
		title := article.Section + ": " + article.Headline
		doc.AddSection(title, article.Body)
	}

	sort.Slice(doc.Sections, func(i, j int) bool {
		return doc.Sections[i].Title < doc.Sections[j].Title
	})

	for doc.Length() > 60000 {
		if len(doc.Sections) == 0 {
			break
		}
		idx := rand.Intn(len(doc.Sections))
		doc.Sections = append(doc.Sections[:idx], doc.Sections[idx+1:]...)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case out <- doc:
	}

	return nil
}
