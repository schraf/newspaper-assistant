package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/schraf/assistant/pkg/eval"
	"github.com/schraf/assistant/pkg/generators"
	"github.com/schraf/assistant/pkg/models"
	"github.com/schraf/newspaper-assistant/internal/newspaper"
	_ "github.com/schraf/newspaper-assistant/pkg/generator"
)

func main() {
	dateRange := flag.String("date_range", "", "Human-readable date range for the edition (required, e.g. \"Nov 28–30, 2025\")")
	location := flag.String("location", "", "Location for the Local section (required, e.g. \"California\")")
	depthString := flag.String("length", "short", "Newspaper length: short, medium, or long (default: short)")
	model := flag.String("model", "", "Model to use for evaluation")
	flag.Parse()

	if *dateRange == "" {
		fmt.Fprintf(os.Stderr, "Error: date_range is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *location == "" {
		fmt.Fprintf(os.Stderr, "Error: location is required\n")
		flag.Usage()
		os.Exit(1)
	}

	var depth newspaper.ResearchDepth
	switch *depthString {
	case "short":
		depth = newspaper.ResearchDepthShort
	case "medium":
		depth = newspaper.ResearchDepthMedium
	case "long":
		depth = newspaper.ResearchDepthLong
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid length '%s'. Must be one of: short, medium, long\n", *depthString)
		os.Exit(1)
	}

	// Create request object
	request := models.ContentRequest{
		Body: map[string]any{
			"date_range":     *dateRange,
			"location":       *location,
			"research_depth": depth,
		},
	}

	ctx := context.Background()

	generator, err := generators.Create("newspaper", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := eval.Evaluate(ctx, generator, request, *model); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
