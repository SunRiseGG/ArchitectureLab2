package gomodule

import (
	"bytes"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"strings"
	"testing"
)

var mockFileSystems = []map[string][]byte{
	{
		"Blueprints": []byte(`
			go_binary {
			  name: "boodjs",
			  pkg: ".",
        testPkg: ".",
			  srcs: ["main_test.go", "main.go",],
			}
		`),
		"main.go": nil,
		"main_test.go": nil,
	},
	{
		"Blueprints": []byte(`
			go_binary {
			  name: "boodjs",
			  pkg: ".",
			  srcs: ["main_test.go", "main.go",],
			}
		`),
		"main.go": nil,
		"main_test.go": nil,
	},

	{
		"Blueprints": []byte(`
			go_binary {
			  name: "boodjs",
			  pkg: ".",
        testPkg: ".",
			  srcs: ["main_test.go", "main.go",],
			  vendorFirst: true
			}
		`),
		"main.go": nil,
		"main_test.go": nil,
	},
}

var testedOutput = [][]string{
	{
		"out/bin/bood/boodjs:",
		"g.gomodule.binaryBuild | main.go\n",
		"out/reports/bood/test.txt",
		"g.gomodule.test | main_test.go main.go",
	},
	{
		"out/bin/bood/boodjs:",
		"g.gomodule.binaryBuild | main.go\n",
	},
	{
		"out/bin/bood/boodjs:",
		"g.gomodule.binaryBuild | main.go vendor\n",
		"build vendor: g.gomodule.vendor | go.mod\n",
		"out/reports/bood/test.txt",
		"g.gomodule.test | main_test.go main.go",
	},
}

func TestSimpleBinFactory(t *testing.T) {
	for index, mockFileSystem := range mockFileSystems {
		t.Run(string(index), func(t *testing.T) {
			ctx := blueprint.NewContext()

			ctx.MockFileSystem(mockFileSystem)

			ctx.RegisterModuleType("go_binary", SimpleBinFactory)

			cfg := bood.NewConfig()

			_, errs := ctx.ParseBlueprintsFiles(".", cfg)
			if len(errs) != 0 {
				t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
			}

			_, errs = ctx.PrepareBuildActions(cfg)
			if len(errs) != 0 {
				t.Errorf("Unexpected errors while preparing build actions: %s", errs)
			}
			buffer := new(bytes.Buffer)
			if err := ctx.WriteBuildFile(buffer); err != nil {
				t.Errorf("Error writing ninja file: %s", err)
			} else {
				text := buffer.String()
				t.Logf("Generated ninja build file:\n%s", text)
				for _, testedOutputString := range testedOutput[index] {
					if !strings.Contains(text, testedOutputString) {
						t.Errorf("Generated ninja file does not have expected string `%s`", testedOutputString)
					}
				}
			}
		})
	}
}
