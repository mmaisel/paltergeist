package engagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pg "palter.io/paltergeist"
	"palter.io/paltergeist/provisioner/pulumi"
)

// Engagement manages a set or related trap resources in the same lifecycle.
type Engagement struct {
	// Id is the unique identifier for the engagement.
	Id uuid.UUID
	// Name of the engagement, human friendly.
	Name        string
	Stratagems  []pg.Stratagem
	provisioner *pulumi.Provisioner
	graph       *pg.Graph
}

func (e *Engagement) String() string {
	return fmt.Sprintf("Engagement{Id: %s, Name: %s}", e.Id, e.Name)
}

// Create a new engagement with the given name.
func New(name string, provisioner *pulumi.Provisioner, stratagems ...pg.Stratagem) *Engagement {
	return &Engagement{
		Id:          uuid.UUID{},
		Name:        name,
		Stratagems:  stratagems,
		provisioner: provisioner,
		graph:       nil,
	}
}

// Graph returns the graph for the engagement.
func (e *Engagement) Graph() *pg.Graph {
	// lazy load the graph
	if e.graph == nil {
		e.graph = e.provisioner.ConstructGraph()
	}
	return e.graph
}

// Plan all stratagems in the engagement and generate and add traps to the graph.
// Each stratagem is executed in sequence with the same graph context.
// If any stratagem returns an error, planning stops and returns the error.
func (e *Engagement) Plan(ctx context.Context) error {
	for _, stratagem := range e.Stratagems {
		err := stratagem(ctx, e.Graph())
		if err != nil {
			return err
		}
	}
	return e.provisioner.Add(e.Name, e.Graph())
}

// Deploy all traps to the environment using the provisioner.
func (e *Engagement) Deploy() error {
	return e.provisioner.Add(e.Name, e.Graph())
}

// Destroy all traps provisioned by the engagement.
func (e *Engagement) Destroy() error {
	return e.provisioner.Destroy(e.Name)
}
