package jsbundlemodule

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
			js_bundle {
			  name: "someBood",
			  srcs: ["foo.js", "bar.js"],
        obfuscate: true,
			}
		`),
			"foo.js": nil,
			"bar.js": nil,
		},
		{
			"Blueprints": []byte(`
			js_bundle {
			  name: "someBood",
			  srcs: ["baz.js", "qux.js"],
        obfuscate: false,
			}
		`),
			"baz.js": nil,
			"qux.js": nil,
		},
	}

var testedOutput = [][]string{
		{
			"build out/js/bood/someBood.js:",
			"g.jsbundlemodule.obfuscate",
			"foo.js bar.js",
		},
		{
			"build out/js/bood/someBood.js:",
			"g.jsbundlemodule.bundle",
			"baz.js qux.js",
		},
	}

func TestSimpleJsFactory(t *testing.T) {
	for index, mockFileSystem := range mockFileSystems {
		t.Run(string(index), func(t *testing.T) {
			ctx := blueprint.NewContext()

			ctx.MockFileSystem(mockFileSystem)

			ctx.RegisterModuleType("js_bundle", SimpleJsFactory)

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
