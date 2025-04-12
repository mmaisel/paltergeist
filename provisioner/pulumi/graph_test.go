package pulumi

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	pg "palter.io/paltergeist"
)

const (
	// XXX: Stack to test export.
	testStack = ""
)

func TestExport(t *testing.T) {
	p, err := New(t.Context(), WithTargetStack(testStack))
	assert.Nil(t, err)
	assert.NotNil(t, p)
	state := p.export(testStack)
	// Save out stack state for debug.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(state); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("state.json", buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

func TestGraph(t *testing.T) {
	p, err := New(t.Context(), WithTargetStack(testStack))
	assert.Nil(t, err)
	assert.NotNil(t, p)
	g := p.ConstructGraph()
	resources := pg.Select[*pg.ServiceAccount](g)
	assert.Len(t, resources, 2)
	users := pg.Select[*pg.User](g)
	assert.Len(t, users, 1)
}
