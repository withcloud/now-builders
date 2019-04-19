package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type analyse struct {
	PackageName string   `json:"packageName"`
	FuncName    string   `json:"functionName"`
	Watch       []string `json:"watch"`
}

func parse(fileName string) *ast.File {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Could not parse Go file \"%s\"\n", fileName)
		os.Exit(1)
	}

	return parsed
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		itf, err := filepath.Match("*test.go", path)
		if err != nil {
			log.Fatal(err)
		}

		// we don't need Dirs, or test files
		// we only want `.go` files
		if info.IsDir() || itf || filepath.Ext(path) != ".go" {
			return nil
		}

		*files = append(*files, path)
		return nil
	}
}

func main() {
	if len(os.Args) != 2 {
		// Args should have the program name on `0`
		// and the file name on `1`
		fmt.Println("Wrong number of args; Usage is:\n  ./analyse file_name.go")
		os.Exit(1)
	}
	fileName := os.Args[1]
	rf, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	se := string(rf)

	var files []string
	var rfs []string

	// looking for all go files that have export func
	// using in entrypoint
	err = filepath.Walk(filepath.Dir(fileName), visit(&files))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// if it isn't entrypoint
		if filepath.Base(fileName) != filepath.Base(file) {
			// find exported func
			pf := parse(file)
			for _, decl := range pf.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					// this declaraction is not a function
					// so we're not interested
					continue
				}
				if fn.Name.IsExported() {

					if strings.Contains(se, fn.Name.Name) {
						// absolute
						// absFile, err := filepath.Abs(file)
						// if err != nil {
						// 	log.Fatal(err)
						// }

						// relative
						rel, err := filepath.Rel(fileName, file)
						if err != nil {
							log.Fatal(err)
						}
						relFile := strings.TrimPrefix(rel, ".")

						rfs = append(rfs, relFile)
					}
				}
			}
		}
	}

	parsed := parse(fileName)
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			// this declaraction is not a function
			// so we're not interested
			continue
		}
		if fn.Name.IsExported() == true {
			// we found the first exported function
			// we're done!
			analysed := analyse{
				PackageName: parsed.Name.Name,
				FuncName:    fn.Name.Name,
				Watch:       rfs,
			}
			json, _ := json.Marshal(analysed)
			fmt.Print(string(json))
			os.Exit(0)
		}
	}
}
