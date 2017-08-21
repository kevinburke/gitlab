package main

import (
	"testing"

	"github.com/pelletier/go-toml"
)

func TestUnmarshal(t *testing.T) {
	in := []byte(`[organizations]

	[organizations.mycompany]
	host = "gitlab.mycompany.com"
	token = "mytoken"
	`)
	c := new(GitlabConfig)
	err := toml.Unmarshal(in, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Organizations) != 1 {
		t.Errorf("wrong number of organizations: %d", len(c.Organizations))
	}
}
