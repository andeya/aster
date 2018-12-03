package examples

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Args simulation code file
type Args struct {
	Filename string
	Src      string
}

var examples = map[string]*Args{
	"structtag": {"../out/eg.structtag.go", `package test
	type S struct {
		Apple string
		BananaPeel,car,OrangeWater string
		E int
	}
	
	func F(){
		type M struct {
			N int
			lowerCase string
		}
	}
	`},
}

var ignored = map[string]bool{}

func TestExamples(t *testing.T) {
	for name, args := range examples {
		if ignored[name] {
			continue
		}
		t.Run(name, func(t *testing.T) {
			testExample(t, name, args)
		})
	}
}

func testExample(t *testing.T, name string, args *Args) {
	p, _ := filepath.Abs(name)
	// filename, _ := filepath.Abs(args.Filename)
	cmd := exec.Command("go", []string{
		"run",
		p,
		"--filename", args.Filename,
		"--src", args.Src,
	}...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Errorf("error running cmd %q", err)
	}
}
