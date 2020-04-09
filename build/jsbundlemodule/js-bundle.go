package jsbundlemodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
	"strings"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/SunRiseGG/ArchitectureLab2/build/jsbundlemodule")

	// Ninja rule to execute js-bundle with obfuscation
	jsObfuscate = pctx.StaticRule("obfuscate", blueprint.RuleParams{
		Command: "cd ${workDir} && npx webpack --mode=production ${input} -o ${output}",
		Description: "Bundle JavaScript files with obfuscation",
	}, "workDir", "input", "output")

	// Ninja rule to execute js-bundle
	jsBundle = pctx.StaticRule("bundle", blueprint.RuleParams{
		Command:     "cd ${workDir} && npx webpack --mode=none ${input} -o ${output}",
		Description: "Bundle JavaScript files",
	}, "workDir", "input", "output")
)

type jsBundleModule struct {
	blueprint.SimpleName

	properties struct {
		Name        string
		Srcs        []string
		SrcsExclude []string
		Obfuscate   bool
	}
}

func (jb *jsBundleModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding bundle actions for js bundle module '%s'", name)

	output := path.Join(config.BaseOutputDir, "js/bood", name+".js")

	var inputs []string
	inputErors := false

	for _, src := range jb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, jb.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	input := strings.Join(inputs[:], " ")

	if jb.properties.Obfuscate && jb.properties.Obfuscate == true {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Bundle JavaScript files with obfuscation %s", name),
			Rule:        jsObfuscate,
			Outputs:     []string{output},
			Implicits:   inputs,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"input":   input,
				"output":  output,
			},
		})

	} else {

		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Bundle JavaScript files %s", name),
			Rule:        jsBundle,
			Outputs:     []string{output},
			Implicits:   inputs,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"input":   input,
				"output":  output,
			},
		})
	}
}

func SimpleJsFactory() (blueprint.Module, []interface{}) {
	mType := &jsBundleModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
