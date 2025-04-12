package paltergeist

import (
	"fmt"
	"log/slog"
)

type Kind string

// Common trap kinds
const (
	KindStorage   Kind = "storage"
	KindCompute   Kind = "compute"
	KindNetwork   Kind = "network"
	KindDatabase  Kind = "database"
	KindSecrets   Kind = "secrets"
	KindContainer Kind = "container"
	KindIdentity  Kind = "identity"
)

// Type of the Resource
type Type int

const (
	ServiceAccountResource Type = iota
	UserResource
	BucketResource
	CloudRunServiceResource
)

// Trap is a deception asset used to detect attackers.
type Trap interface {
	// TrapId is the unique identifier for the trap.
	TrapId() string
	// Kind of the trap, used to group related traps together.
	Kind() Kind
	// Type of the trap, maps to specific types of resource.
	Type() Type
}

type Resource struct {
	// IsTrap marks if this resource is a Trap.
	isTrap bool
	value  Trap
}

// IsTrap returns true if the resource is a trap.
func (r Resource) IsTrap() bool {
	return r.isTrap
}

// Id returns the Id number of the Resource, used an internal identifier.
func (r Resource) Id() string {
	// Just return the TrapId for now.
	return r.value.TrapId()
}

// Graph of Resources and Traps!
type Graph struct {
	nodes map[string]Resource
}

// Resources returns all resources in the graph.
func (g *Graph) Resources() []Resource {
	var resources []Resource
	for _, r := range g.nodes {
		resources = append(resources, r)
	}
	return resources
}

// Traps returns all traps in the graph.
func (g *Graph) Traps() []Trap {
	var traps []Trap
	for _, r := range g.nodes {
		if r.IsTrap() {
			traps = append(traps, r.value)
		}
	}
	return traps
}

// Add a Resource to the graph.
func (g *Graph) Add(resource Trap) error {
	r := Resource{
		isTrap: false,
		value:  resource,
	}
	if _, ok := g.nodes[r.Id()]; ok {
		return fmt.Errorf("resource with id %s already exists", r.Id())
	}
	g.nodes[r.Id()] = r
	slog.Info("added", "resource", resource)
	return nil
}

// AddTrap adds a Resource marked as a trap to the graph.
func (g *Graph) AddTrap(trap Trap) error {
	r := Resource{
		isTrap: true,
		value:  trap,
	}
	if _, ok := g.nodes[r.Id()]; ok {
		return fmt.Errorf("trap with id %s already exists", r.Id())
	}
	g.nodes[r.Id()] = r
	slog.Info("added trap", "trap", trap)
	return nil
}

// New constructs an empty graph.
func New() *Graph {
	return &Graph{
		nodes: make(map[string]Resource),
	}
}

type Predicate func(R Resource) bool

func IsTrap(r Resource) bool {
	return r.IsTrap()
}

func IsTarget(r Resource) bool {
	return !r.IsTrap()
}

// Select all resources of a given type.
func Select[T Trap](g *Graph, predicates ...Predicate) []T {
	var resources []T
	for _, r := range g.nodes {
		if value, ok := r.value.(T); ok {
			// Check if the resource matches all predicates
			matches := true
			for _, predicate := range predicates {
				if !predicate(r) {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
			resources = append(resources, value)
		}
	}
	return resources
}
