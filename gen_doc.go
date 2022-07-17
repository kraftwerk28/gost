package main

import (
	"flag"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Version string

type blockletField struct {
	Name, DataType string
	Doc            string
	Since          Version
}

type BlockletDoc struct {
	FileName     string
	Name         string
	Doc          string
	ConfigFields []blockletField
	Since        Version
}

type Template struct {
	Blocklets []BlockletDoc
}

var pascalRe = regexp.MustCompile(`\B[A-Z]`)

func structNameToBlockletName(s string) string {
	if strings.HasSuffix(s, "BlockConfig") {
		s = strings.TrimSuffix(s, "BlockConfig")
	} else {
		s = strings.TrimSuffix(s, "Config")
	}
	s = pascalRe.ReplaceAllStringFunc(s, func(m string) string {
		return "_" + m
	})
	return strings.ToLower(s)
}

func getFieldTypeName(e ast.Expr) string {
	if ee, ok := e.(*ast.Ident); ok {
		return ee.Name
	}
	if ee, ok := e.(*ast.StarExpr); ok {
		return ee.X.(*ast.Ident).Name
	}
	return ""
}

func escapeTableDelimiters(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

var tagRe = regexp.MustCompile(`yaml:"([^"]+)"`)

func tagToFieldName(tag string) string {
	m := tagRe.FindStringSubmatch(tag)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func main() {
	tmplFileFlag := flag.String("t", "doc/template.md", "Markdown file with template")
	outFileFlag := flag.String("o", "README.md", "Output file")
	flag.Parse()
	fSet := token.NewFileSet()
	blocklets := make([]BlockletDoc, 0)

	for _, p := range flag.Args() {
		log.Printf("Processing %s\n", p)
		astFile, err := parser.ParseFile(fSet, p, nil, parser.ParseComments)
		if err != nil {
			log.Print(err)
			continue
		}
		pkg, err := doc.NewFromFiles(fSet, []*ast.File{astFile}, "")
		if err != nil {
			log.Print(err)
			continue
		}
		var cfgTyp *doc.Type
		for _, typ := range pkg.Types {
			if strings.HasSuffix(typ.Name, "Config") {
				cfgTyp = typ
				break
			}
		}
		if cfgTyp == nil {
			continue
		}
		fields := make([]blockletField, 0)
		st := cfgTyp.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
		for _, f := range st.Fields.List {
			if f.Names == nil {
				continue
			}
			var dataType string
			if t, ok := f.Type.(*ast.StarExpr); ok {
				dataType = t.X.(*ast.Ident).String()
			} else if t, ok := f.Type.(*ast.Ident); ok {
				dataType = t.Name
			} else {
				continue
			}
			fieldName := tagToFieldName(f.Tag.Value)
			fields = append(fields, blockletField{
				Name:     fieldName,
				DataType: dataType,
				Doc:      strings.TrimSpace(f.Doc.Text()),
			})
		}
		blocklets = append(blocklets, BlockletDoc{
			Name:         structNameToBlockletName(cfgTyp.Name),
			FileName:     p,
			Doc:          cfgTyp.Doc,
			ConfigFields: fields,
		})
	}

	var outFile *os.File
	if *outFileFlag == "-" {
		outFile = os.Stdout
	} else {
		f, err := os.Create(*outFileFlag)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		outFile = f
	}

	tmplData := Template{blocklets}
	funcMap := template.FuncMap{
		"escapeTableDelimiters": escapeTableDelimiters,
	}
	t := template.Must(
		template.New(filepath.Base(*tmplFileFlag)).Funcs(funcMap).ParseFiles(*tmplFileFlag),
	)
	if err := t.Execute(outFile, tmplData); err != nil {
		panic(err)
	}
}
