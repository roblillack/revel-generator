package main

import (
	"flag"
	"fmt"
	"github.com/revel/revel"
	"github.com/revel/revel/harness"
	"os"
	"path"
	"text/template"
)

func main() {
	targetFile := flag.String("t", "app/setup.go", "Target file to generate.")
	routeTarget := flag.String("r", "app/routes/routes.go", "Target route file to generate.")
	runMode := flag.String("m", "", "RunMode to initialize application in.")
	flag.Parse()

	packageName := flag.Arg(0)
	if packageName == "" {
		fmt.Fprintln(os.Stderr, "Please specify your application's import path as argument!")
		return
	}

	revel.Init(*runMode, packageName, "")

	fmt.Printf("Generating %s and %s ...\n", *targetFile, *routeTarget)

	sourceInfo, compileError := harness.ProcessSource(revel.CodePaths)
	if compileError != nil {
		panic(compileError)
	}

	// Add the db.import to the import paths.
	if dbImportPath, found := revel.Config.String("db.import"); found {
		sourceInfo.InitImportPaths = append(sourceInfo.InitImportPaths, dbImportPath)
	}

	// Generate two source files.
	templateArgs := map[string]interface{}{
		"Controllers":    sourceInfo.ControllerSpecs(),
		"ValidationKeys": sourceInfo.ValidationKeys,
		"ImportPaths":    calcImportAliases(sourceInfo),
		"TestSuites":     sourceInfo.TestSuites(),
	}

	generateSources(SETUP, *targetFile, templateArgs)
	generateSources(harness.ROUTES, *routeTarget, templateArgs)
}

func generateSources(tpl, filename string, templateArgs map[string]interface{}) {
	sourceCode := revel.ExecuteTemplate(template.Must(template.New("").Parse(tpl)), templateArgs)

	os.MkdirAll(path.Dir(filename), 0755)

	// Create the file
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		revel.ERROR.Fatalf("Failed to create file: %v", err)
	}
	_, err = file.WriteString(sourceCode)
	if err != nil {
		revel.ERROR.Fatalf("Failed to write to file: %v", err)
	}
}

// Looks through all the method args and returns a set of unique import paths
// that cover all the method arg types.
// Additionally, assign package aliases when necessary to resolve ambiguity.
func calcImportAliases(src *harness.SourceInfo) map[string]string {
	aliases := make(map[string]string)
	typeArrays := [][]*harness.TypeInfo{src.ControllerSpecs(), src.TestSuites()}
	for _, specs := range typeArrays {
		for _, spec := range specs {
			addAlias(aliases, spec.ImportPath, spec.PackageName)

			for _, methSpec := range spec.MethodSpecs {
				for _, methArg := range methSpec.Args {
					if methArg.ImportPath == "" {
						continue
					}

					addAlias(aliases, methArg.ImportPath, methArg.TypeExpr.PkgName)
				}
			}
		}
	}

	// Add the "InitImportPaths", with alias "_"
	/*for _, importPath := range src.InitImportPaths {
		if _, ok := aliases[importPath]; !ok {
			aliases[importPath] = "_"
		}
	}*/

	return aliases
}

func addAlias(aliases map[string]string, importPath, pkgName string) {
	alias, ok := aliases[importPath]
	if ok {
		return
	}
	alias = makePackageAlias(aliases, pkgName)
	aliases[importPath] = alias
}

func makePackageAlias(aliases map[string]string, pkgName string) string {
	i := 0
	alias := pkgName
	for containsValue(aliases, alias) {
		alias = fmt.Sprintf("%s%d", pkgName, i)
		i++
	}
	return alias
}

func containsValue(m map[string]string, val string) bool {
	for _, v := range m {
		if v == val {
			return true
		}
	}
	return false
}

const SETUP = `// GENERATED CODE - DO NOT EDIT
package app

import (
	"reflect"
	"github.com/revel/revel"{{range $k, $v := $.ImportPaths}}
	{{$v}} "{{$k}}"{{end}}
)

var (
	// So compiler won't complain if the generated code doesn't reference reflect package...
	_ = reflect.Invalid
)

func SetupRevel() {
	{{range $i, $c := .Controllers}}
	revel.RegisterController((*{{index $.ImportPaths .ImportPath}}.{{.StructName}})(nil),
		[]*revel.MethodType{
			{{range .MethodSpecs}}&revel.MethodType{
				Name: "{{.Name}}",
				Args: []*revel.MethodArg{ {{range .Args}}
					&revel.MethodArg{Name: "{{.Name}}", Type: reflect.TypeOf((*{{index $.ImportPaths .ImportPath | .TypeExpr.TypeName}})(nil)) },{{end}}
				},
				RenderArgNames: map[int][]string{ {{range .RenderCalls}}
					{{.Line}}: []string{ {{range .Names}}
						"{{.}}",{{end}}
					},{{end}}
				},
			},
			{{end}}
		})
	{{end}}
	revel.DefaultValidationKeys = map[string]map[int]string{ {{range $path, $lines := .ValidationKeys}}
		"{{$path}}": { {{range $line, $key := $lines}}
			{{$line}}: "{{$key}}",{{end}}
		},{{end}}
	}
	revel.TestSuites = []interface{}{ {{range .TestSuites}}
		(*{{index $.ImportPaths .ImportPath}}.{{.StructName}})(nil),{{end}}
	}
}
`
