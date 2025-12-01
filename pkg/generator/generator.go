package generator

import (
	"context"
	"fmt"

	"github.com/schraf/assistant/pkg/generators"
	"github.com/schraf/assistant/pkg/models"
	"github.com/schraf/newspaper-assistant/internal/newspaper"
)

func init() {
	generators.MustRegister("newspaper", factory)
}

func factory(generators.Config) (models.ContentGenerator, error) {
	return &generator{}, nil
}

type generator struct{}

func (g *generator) Generate(ctx context.Context, request models.ContentRequest, assistant models.Assistant) (*models.Document, error) {
	var depth newspaper.ResearchDepth
	switch v := request.Body["research_depth"].(type) {
	case newspaper.ResearchDepth:
		depth = v
	case int:
		depth = newspaper.ResearchDepth(v)
	case int64:
		depth = newspaper.ResearchDepth(v)
	case float64:
		depth = newspaper.ResearchDepth(v)
	default:
		return nil, fmt.Errorf("no research depth (expected short/medium/long mapping)")
	}

	if !depth.Validate() {
		return nil, fmt.Errorf("invalid research depth")
	}

	dateRange, _ := request.Body["date_range"].(string)
	location, _ := request.Body["location"].(string)

	opts := newspaper.NewspaperOptions{
		DateRange: dateRange,
		Location:  location,
		Depth:     depth,
	}

	return newspaper.GenerateNewspaper(ctx, assistant, opts)
}
