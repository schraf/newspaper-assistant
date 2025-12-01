# Newspaper Assistant

A content generator plugin for the [assistant](https://github.com/schraf/assistant) project that produces multi-section newspaper-style documents over a specified date range.

## Overview

This project provides a **newspaper** content generator that integrates with the assistant framework. It plans sections and articles, performs web-based research, and synthesizes a complete newspaper edition with Local, US, World, Business and Financial, Technology, and Health and Science sections.

## Features

- **Newspaper Generator**: Implements the `ContentGenerator` interface from the assistant project under the name `newspaper`.
- **Configurable Length**: Supports three edition sizes (`short`, `medium`, `long`) which control how many articles appear per section.
- **Section-Aware Planning**: Plans articles separately for each fixed section using a two-step ideas → structured plan flow.
- **Iterative Research**: Uses iterative research and analysis to gather facts for each planned article.
- **Standalone CLI**: Can be run as a standalone command-line tool for testing.

## Usage

### As a Plugin

This project is imported by the assistant project and automatically registers the `"newspaper"` generator. The generator can be invoked through the assistant's API endpoints by specifying the generator name and providing:

- `date_range` – human-readable date window for the edition (e.g. `"Nov 28–30, 2025"`).
- `location` – location used for the Local section (e.g. `"California"`).
- `research_depth` – integer corresponding to `short`/`medium`/`long` (0, 1, 2).

### As a Standalone Tool

Build and run the CLI tool:

```bash
make build
./newspaper -date_range "Nov 28–30, 2025" -location "California" -length medium
```

Length options:
- `short` – 3 articles per section
- `medium` – 5 articles per section  
- `long` – 8 articles per section

## Development

```bash
# Build the project
make build

# Run tests
make test

# Format code
make fmt

# Run vetting
make vet
```

## Project Structure

- `cmd/main.go` - Standalone CLI application
- `pkg/generator/` - Generator plugin implementation (`newspaper` generator)
- `internal/newspaper/` - Core newspaper planning, research, synthesis, and editing

## Dependencies

- `github.com/schraf/assistant` - Assistant framework
- Google Cloud APIs (via `google.golang.org/genai`)
