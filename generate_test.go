package paltergeist

import (
	"log/slog"
	"os"
	"testing"

	"palter.io/paltergeist/generator"

	"github.com/stretchr/testify/assert"
)

var (
	samples = []ServiceAccount{
		{
			Id:          "wlifgke-backend-sa",
			Name:        "Wlifgke Backend Service Account",
			Description: "Service Account for Cloud Run",
			Email:       "wlifgke-backend-sa@wlifgke.iam.gserviceaccount.com",
		},
		{
			Id:          "vm-travelexpenses-sa",
			Name:        "Travel expenses VM Service Account",
			Description: "Service Account for Cloud Run",
			Email:       "vm-travelexpenses-sa@wlifgke.iam.gserviceaccount.com",
		},
	}
	projectId string
)

func init() {
	projectId = os.Getenv("PALTERGEIST_PROJECT_ID")
}

func TestPrompts(t *testing.T) {
	traps := make([]Trap, len(samples))
	for i, t := range samples {
		traps[i] = &t
	}
	prompt, err := generator.PromptTrapGeneration(traps)
	assert.Nil(t, err)
	assert.NotEmpty(t, prompt)
	slog.Info(prompt)
}

func TestGenerateTrap(t *testing.T) {
	generator.Initialize(t.Context(), projectId)
	traps, err := generator.GenerateTraps(t.Context(), samples, 3)
	for _, trap := range traps {
		assert.Nil(t, err)
		assert.NotNil(t, trap)
		assert.NotEqual(t, trap.Id, "")
		assert.NotEqual(t, trap.Name, "")
		assert.NotEqual(t, trap.Description, "")
		assert.NotEqual(t, trap.Email, "")
		slog.Info("trap", "trap", trap, "err", err)
	}
}
