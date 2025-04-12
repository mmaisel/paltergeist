package generator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"cloud.google.com/go/vertexai/genai"
)

var (
	client *genai.Client
)

// Initialize generator resources
func Initialize(ctx context.Context, projectId string) {
	var err error
	client, err = genai.NewClient(ctx, projectId, "us-central1")
	if err != nil {
		panic(fmt.Errorf("unable to create client: %v", err))
	}
}

func Close() {
	if err := client.Close(); err != nil {
		panic(err)
	}
}

// GenerateTraps generates n new traps based on the provided examples.
// It calls GenerateTrap n times to create unique trap instances.
//
// Parameters:
//   - examples: A slice of type T instances used as training examples for the AI model
//   - n: The number of traps to generate
//
// Returns:
//   - A slice of newly generated trap instances
//   - An error if any step of the generation process fails
func GenerateTraps[T any](ctx context.Context, examples []T, n int) ([]T, error) {
	traps := make([]T, 0, n)
	for i := 0; i < n; i++ {
		trap, err := GenerateTrap(ctx, examples)
		if err != nil {
			return nil, fmt.Errorf("failed to generate trap %d: %v", i, err)
		}
		traps = append(traps, *trap)
	}
	return traps, nil
}

// GenerateTrap uses a generative AI model to generate a new instance of type T based on provided examples.
// It feeds the examples into a prompt, sends it to Google's Gemini AI model,
// and returns a new instance conforming to the schema of type T.
//
// Parameters:
//   - examples: A slice of type T instances used as in-context examples for the LLM.
//
// Returns:
//   - A pointer to a newly generated instance of type T
//   - An error if any step of the generation process fails
func GenerateTrap[T any](ctx context.Context, examples []T) (*T, error) {

	if client == nil {
		return nil, errors.New("client is not initialized")
	}

	// Create an empty instance of T to derive its schema
	var t T

	// Convert type T to a GenAI schema that defines the expected schema structure
	schema := IntoSchema(t)

	// Convert the schema to JSON for logging and debugging
	// schemaJson, _ := json.MarshalIndent(schema, "", "  ")
	// Write the schema to a file for debugging purposes
	//if err := os.WriteFile("schema.json", schemaJson, 0644); err != nil {
	//	log.Fatal(err)
	//}

	// Configure the generative model with the appropriate response schema
	model := client.GenerativeModel("gemini-2.0-flash-lite")
	model.GenerationConfig.ResponseMIMEType = "application/json"
	model.ResponseSchema = &schema
	// Generate the prompt text using the provided examples
	prompt, err := PromptTrapGeneration(examples)
	if err != nil {
		return nil, err
	}
	slog.Debug("prompt", "prompt", prompt)

	// Send the prompt to the AI model to generate content
	res, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("unable to generate contents: %v", err)
	}

	// Extract the generated content from the response
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if _, err = fmt.Fprint(writer, res.Candidates[0].Content.Parts[0]); err != nil {
		return nil, err
	}
	if err = writer.Flush(); err != nil {
		return nil, err
	}

	// Parse the JSON response into the type T
	err = json.Unmarshal(b.Bytes(), &t)
	return &t, err
}
