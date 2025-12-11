package generator

import (
	"context"
	"os"
	"testing"

	"github.com/schraf/assistant/pkg/eval"
	"github.com/schraf/assistant/pkg/generators"
	"github.com/schraf/assistant/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	os.Setenv("ASSISTANT_PROVIDER", "mock")

	request := models.ContentRequest{
		Body: map[string]any{
			"days_back": 7,
			"location":  "North Caolina, USA",
		},
	}

	ctx := context.Background()

	generator, err := generators.Create("newspaper", nil)
	require.NoError(t, err)

	err = eval.Evaluate(ctx, generator, request, "mock-model")
	assert.NoError(t, err)
}
