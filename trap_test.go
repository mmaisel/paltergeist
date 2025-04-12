package paltergeist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph(t *testing.T) {
	g := New()
	sa := ServiceAccount{
		Id:          "foo",
		Name:        "foo",
		Description: "",
		Email:       "",
	}
	err := g.Add(&sa)
	assert.Nil(t, err)

	trap := ServiceAccount{
		Id:          "bar",
		Name:        "bar",
		Description: "",
		Email:       "",
	}
	err = g.AddTrap(&trap)
	assert.Nil(t, err)

	serviceAccounts := Select[*ServiceAccount](g)
	for _, sa := range serviceAccounts {
		t.Log(sa)
	}
	assert.Len(t, serviceAccounts, 2)
}
