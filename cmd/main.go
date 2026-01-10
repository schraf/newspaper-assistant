package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/schraf/assistant/pkg/eval"
	"github.com/schraf/assistant/pkg/generators"
	"github.com/schraf/assistant/pkg/models"
	_ "github.com/schraf/newspaper-assistant/pkg/generator"
)

func main() {
	daysBack := flag.Int("days_back", 1, "Number of days in the past to include (e.g. 3 means from 3 days ago through today)")
	maxLength := flag.Int("max_length", 60000, "Max legnth of newspaper document")
	location := flag.String("location", "", "Location for the Local section (required, e.g. \"California\")")
	flag.Parse()

	if *daysBack <= 0 {
		fmt.Fprintf(os.Stderr, "Error: days_back must be a positive integer\n")
		flag.Usage()
		os.Exit(1)
	}

	if *maxLength <= 0 {
		fmt.Fprintf(os.Stderr, "Error: max_length must be a positive integer\n")
		flag.Usage()
		os.Exit(1)
	}

	if *location == "" {
		fmt.Fprintf(os.Stderr, "Error: location is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create request object
	request := models.ContentRequest{
		Body: map[string]any{
			"days_back":  *daysBack,
			"max_length": *maxLength,
			"location":   *location,
		},
	}

	ctx := context.Background()

	generator, err := generators.Create("newspaper", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := eval.Evaluate(ctx, generator, request, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
