package newspaper

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/schraf/assistant/pkg/models"
)

// GenerateNewspaper orchestrates planning, researching, synthesizing, and editing
// a complete newspaper-style document for the given options.
func GenerateNewspaper(ctx context.Context, assistant models.Assistant, opts NewspaperOptions) (*models.Document, error) {
	slog.InfoContext(ctx, "generating_newspaper",
		slog.Int("days_back", opts.DaysBack),
		slog.String("location", opts.Location),
	)

	plan, err := GenerateNewspaperPlan(ctx, assistant, opts)
	if err != nil {
		return nil, err
	}

	sectionResearch := make([]SectionResearch, 0, len(plan.Sections))

	for _, section := range plan.Sections {
		sectionRes := SectionResearch{
			Plan:     section,
			Articles: make([]ArticleResearch, len(section.ArticlePlans)),
		}

		var wg sync.WaitGroup

		for i, articlePlan := range section.ArticlePlans {
			i := i
			articlePlan := articlePlan

			wg.Add(1)

			go func() {
				defer wg.Done()

				slog.Info("starting_article_research",
					slog.String("section", string(section.Type)),
					slog.String("headline", articlePlan.Headline),
				)

				sectionRes.Articles[i] = ResearchArticle(ctx, assistant, section, articlePlan, opts)

				slog.Info("finished_article_research",
					slog.String("section", string(section.Type)),
					slog.String("headline", articlePlan.Headline),
				)
			}()
		}

		wg.Wait()

		sectionResearch = append(sectionResearch, sectionRes)
	}

	doc, err := SynthesizeNewspaper(ctx, assistant, opts, sectionResearch)
	if err != nil {
		return nil, err
	}

	doc, err = EditNewspaper(ctx, assistant, doc)
	if err != nil {
		return nil, err
	}

	doc.Title = "News Report"
	doc.Author = os.Getenv("TELEGRAPH_AUTHOR_NAME")

	return doc, nil
}

// ResearchArticle gathers knowledge for a single planned article, using an
// iterative ask/analyze loop similar to the legacy ResearchSubTopic flow.
func ResearchArticle(ctx context.Context, assistant models.Assistant, section SectionPlan, articlePlan ArticlePlan, opts NewspaperOptions) ArticleResearch {
	result := ArticleResearch{
		Plan:      articlePlan,
		Knowledge: []Knowledge{},
	}

	maxIterations := 0

	switch opts.Depth {
	case ResearchDepthShort:
		maxIterations = 0
	case ResearchDepthMedium:
		maxIterations = 2
	case ResearchDepthLong:
		maxIterations = 5
	}

	goal := "Write a complete, accurate, and balanced news article for the " +
		string(section.Type) + " section based on this story."

	// Initial questions come from the plan; they may be refined by analysis.
	questions := append([]string(nil), articlePlan.Questions...)

	// Build initial knowledge.
	initialKnowledge, err := GenerateArticleKnowledge(ctx, assistant, section, articlePlan, opts, questions)
	if err != nil {
		result.Error = err
		return result
	}

	result.Knowledge = append(result.Knowledge, initialKnowledge...)

	for iteration := 0; iteration < maxIterations; iteration++ {
		questions, err = AnalyzeArticleKnowledge(ctx, assistant, goal, section, articlePlan, questions, result.Knowledge, opts)
		if err != nil {
			result.Error = err
			break
		}

		if len(questions) == 0 {
			break
		}

		newKnowledge, err := GenerateArticleKnowledge(ctx, assistant, section, articlePlan, opts, questions)
		if err != nil {
			result.Error = err
			break
		}

		result.Knowledge = append(result.Knowledge, newKnowledge...)
	}

	return result
}
