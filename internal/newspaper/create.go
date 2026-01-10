package newspaper

import (
	"context"
	"fmt"

	"github.com/schraf/assistant/pkg/models"
	"github.com/schraf/pipeline"
)

func CreateNewspaper(ctx context.Context, assistant models.Assistant, section Section, options NewspaperOptions) (*models.Document, error) {
	//--===============================================================--
	//--== CREATE PIPELINE
	//--===============================================================--

	ctx = withAssistant(ctx, assistant)
	ctx = withOptions(ctx, options)
	pipe, ctx := pipeline.WithPipeline(ctx)

	//--===============================================================--
	//--== SET PIPELINE CAPACITY
	//--===============================================================--

	// channel capacity size
	const capacity = 2

	//--===============================================================--
	//--== STAGE 0 : SOURCE NEWSPAPER SECTION
	//--===============================================================--

	stage0 := make(chan Section, 1)
	stage0 <- section
	close(stage0)

	//--===============================================================--
	//--== STAGE 1 : PLAN ARTICLES FOR SECTION
	//--===============================================================--

	stage1 := make(chan []Article, capacity)
	pipeline.Transform(pipe, Plan, stage0, stage1)

	//--===============================================================--
	//--== STAGE 2 : FLATTEN ALL ARTICLES
	//--===============================================================--

	stage2 := make(chan Article, capacity)
	pipeline.Flatten(pipe, stage1, stage2)

	//--===============================================================--
	//--== STAGE 3 : RESEARCH EACH ARTICLE
	//--===============================================================--

	stage3 := make(chan Article, capacity)
	pipeline.Transform(pipe, ResearchArticle, stage2, stage3)

	//--===============================================================--
	//--== STAGE 4 : FILTER OUT ANY INVALID ARTICLES
	//--===============================================================--

	stage4 := make(chan Article, capacity)
	pipeline.Filter(pipe, filterValidArticles, stage3, stage4)

	//--===============================================================--
	//--== STAGE 5 : SYNTHESIZE THE RESEARCH
	//--===============================================================--

	stage5 := make(chan Article, capacity)
	pipeline.Transform(pipe, SynthesizeArticle, stage4, stage5)

	//--===============================================================--
	//--== STAGE 6 : FILTER OUT ANY INVALID ARTICLES
	//--===============================================================--

	stage6 := make(chan Article, capacity)
	pipeline.Filter(pipe, filterValidArticles, stage5, stage6)

	//--===============================================================--
	//--== STAGE 7 : AGGREGATE ALL ARTICLES
	//--===============================================================--

	stage7 := make(chan []Article, 1)
	pipeline.Aggregate(pipe, stage6, stage7)

	//--===============================================================--
	//--== STAGE 8 : EDIT FINAL NEWSPAPER
	//--===============================================================--

	stage8 := make(chan models.Document, 1)
	pipeline.Transform(pipe, EditNewspaper, stage7, stage8)

	//--===============================================================--
	//--== GET NEWSPAPER
	//--===============================================================--

	if err := pipe.Wait(); err != nil {
		return nil, fmt.Errorf("failed during newspaper pipeline: %w", err)
	}

	newspaper := <-stage8

	return &newspaper, nil
}
