package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
	"regexp"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/SunRiseGG/ArchitectureLab2/build/gomodule")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd ${workDir} && go build -o ${output} ${pkg}",
		Description: "build go command ${pkg}",
	}, "workDir", "output", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd ${workDir} && go mod vendor",
		Description: "vendor dependencies of ${name}",
	}, "workDir", "name")

	// Ninja rule to execute go test.
	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command:     "cd ${workDir}  && go test -v ${pkg} > ${testOutput}",
		Description: "test ${pkg}",
	}, "workDir", "pkg", "testOutput")
)

type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Name        string
		Pkg         string
		TestPkg     string
		Srcs        []string
		SrcsExclude []string
		VendorFirst bool
	}
}


func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	output := path.Join(config.BaseOutputDir, "bin/bood", name)
	testOutput := path.Join(config.BaseOutputDir, "reports/bood", "test.txt")

	var inputs []string
	var testInputs []string
	inputErrors := false

	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.SrcsExclude); err == nil {
			testInputs = append(testInputs, matches...)
			for _, i := range matches {
				if val, _ := regexp.Match("^.*_test.go$", []byte(i)); val == false {
					inputs = append(inputs, i)
				}
			}
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}
	if inputErrors {
		return
	}

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},

		})

		inputs = append(inputs, vendorDirPath)
		testInputs = append(testInputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{output},
		Implicits:   inputs,

		Args: map[string]string{
			"output": output,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.Pkg,
		},
	})

	if len(tb.properties.TestPkg) > 0 {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Initiate %s tests to Go binary", name),
			Rule:        goTest,
			Outputs:     []string{testOutput},
			Implicits:   testInputs,
			Args: map[string]string{
				"testOutput": testOutput,
				"workDir":    ctx.ModuleDir(),
				"pkg":        tb.properties.TestPkg,
			},
		})
	}

}

func SimpleBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
