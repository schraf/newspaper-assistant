package newspaper

import (
	"bytes"
	"fmt"
	"text/template"
)

type PromptArgs map[string]any

func BuildPrompt(promptTemplate string, args any) (*string, error) {
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, args)
	if err != nil {
		return nil, fmt.Errorf("failed executing prompt template: %w", err)
	}

	prompt := buffer.String()

	return &prompt, nil
}
