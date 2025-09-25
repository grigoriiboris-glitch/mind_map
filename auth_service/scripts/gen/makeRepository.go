package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type Field struct {
	Name string
	DB   string
}

type TemplateData struct {
	PackagePath        string
	StructName         string
	TableName          string
	PrimaryKey         string
	PrimaryKeyField    string
	InsertColumns      string
	InsertPlaceholders string
	InsertFields       string
	SelectColumns      string
	ScanFields         string
	UpdateAssignments  string
	UpdateFields       string
	UpdateLastIndex    int
}

func main() {
	model := flag.String("model", "", "Имя структуры (например Log)")
	table := flag.String("table", "", "Имя таблицы (например logs)")
	out := flag.String("out", "", "Файл для генерации")
	modelsDir := flag.String("", "./models", "Путь к пакету моделей")
	pkgPath := flag.String("pkg", "github.com/mymindmap/api/models", "Импортный путь пакета моделей")
	tmplPath := flag.String("tmpl", "/app/scripts/repository.tmpl", "Путь к шаблону")
	flag.Parse()

	if *out == "" {
    fileName := fmt.Sprintf("%s_repo.go", toSnakeCase(*model))
    *out = filepath.Join("repository", fileName)
    }

	// парсим директорию с моделями
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, *modelsDir, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	var fields []Field
	found := false

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if ts.Name.Name == *model {
						st, ok := ts.Type.(*ast.StructType)
						if !ok {
							continue
						}
						for _, field := range st.Fields.List {
							if field.Tag != nil && len(field.Names) > 0 {
								tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
								if dbTag := tag.Get("db"); dbTag != "" {
									fields = append(fields, Field{
										Name: field.Names[0].Name,
										DB:   dbTag,
									})
								}
							}
						}
						found = true
					}
				}
			}
		}
	}

	if !found {
		panic("Модель не найдена: " + *model)
	}
	if len(fields) == 0 {
		panic("У модели нет полей с тегами db: " + *model)
	}

	// первый field = primary key
	pk := fields[0]

	insertCols := []string{}
	insertPh := []string{}
	insertFields := []string{}
	updateAssignments := []string{}
	updateFields := []string{}
	selectCols := []string{}
	scanFields := []string{}

	i := 1
	for _, f := range fields {
		if f.DB == pk.DB {
			continue
		}
		insertCols = append(insertCols, f.DB)
		insertPh = append(insertPh, fmt.Sprintf("$%d", i))
		insertFields = append(insertFields, "entity."+f.Name)
		updateAssignments = append(updateAssignments, fmt.Sprintf("%s = $%d", f.DB, i))
		updateFields = append(updateFields, "entity."+f.Name)
		i++
	}
	updateFields = append(updateFields, "entity."+pk.Name)

	for _, f := range fields {
		selectCols = append(selectCols, f.DB)
		scanFields = append(scanFields, "&entity."+f.Name)
	}

	data := TemplateData{
		PackagePath:        *pkgPath,
		StructName:         *model,
		TableName:          *table,
		PrimaryKey:         pk.DB,
		PrimaryKeyField:    pk.Name,
		InsertColumns:      strings.Join(insertCols, ", "),
		InsertPlaceholders: strings.Join(insertPh, ", "),
		InsertFields:       strings.Join(insertFields, ", "),
		SelectColumns:      strings.Join(selectCols, ", "),
		ScanFields:         strings.Join(scanFields, ", "),
		UpdateAssignments:  strings.Join(updateAssignments, ", "),
		UpdateFields:       strings.Join(updateFields, ", "),
		UpdateLastIndex:    len(updateFields),
	}

	// читаем шаблон
	tmplBytes, err := os.ReadFile(*tmplPath)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("_repo").Parse(string(tmplBytes))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0755); err != nil {
		panic(err)
	}

	if err := os.WriteFile(*out, buf.Bytes(), 0644); err != nil {
		panic(err)
	}

	fmt.Println("Сгенерирован репозиторий:", *out)

	// создаём тест
testFileName := strings.Replace(filepath.Base(*out), ".go", "_test.go", 1)
testPath := filepath.Join(filepath.Dir(*out), testFileName)

testTmplBytes, err := os.ReadFile("/app/scripts/repository_test.tmpl")
if err != nil {
    panic(err)
}
testTmpl, err := template.New("repo_test").Parse(string(testTmplBytes))
if err != nil {
    panic(err)
}
var testBuf bytes.Buffer
if err := testTmpl.Execute(&testBuf, data); err != nil {
    panic(err)
}
if err := os.WriteFile(testPath, testBuf.Bytes(), 0644); err != nil {
    panic(err)
}

fmt.Println("Сгенерирован репозиторий:", *out)
fmt.Println("Сгенерирован тест:", testPath)

}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}