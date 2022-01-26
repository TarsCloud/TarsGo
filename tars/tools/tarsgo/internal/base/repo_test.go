package base

import (
	"context"
	"testing"
)

func TestRepo(t *testing.T) {
	r := NewRepo("https://github.com/TarsCloud/TarsGo.git", "", "")
	if err := r.Clone(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.CopyTo(context.Background(), &Project{
		App:          "DemoApp",
		Server:       "DemoServer",
		Servant:      "DemoServant",
		GoModuleName: "demo",
	}, "github.com/TarsCloud/TarsGo", "", []string{}); err != nil {
		t.Fatal(err)
	}
}
