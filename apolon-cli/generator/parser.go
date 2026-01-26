package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Parser handles AST parsing of Go source files
type Parser struct {
	inputDir string
}

// NewParser creates a new parser for the given directory
func NewParser(inputDir string) *Parser {
	return &Parser{inputDir: inputDir}
}

// Parse parses all Go files in the directory and returns models grouped by source file
func (p *Parser) Parse() (map[string][]ModelInfo, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, p.inputDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_fields.go") &&
			!strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]ModelInfo)
	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			models := p.extractModels(file, pkg.Name)
			if len(models) > 0 {
				result[filename] = models
			}
		}
	}

	return result, nil
}

// extractModels extracts model information from an AST file
func (p *Parser) extractModels(file *ast.File, pkgName string) []ModelInfo {
	var models []ModelInfo

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		if !p.hasDbTag(structType) {
			return true
		}

		model := ModelInfo{
			Name:    typeSpec.Name.Name,
			Table:   strings.ToLower(typeSpec.Name.Name) + "s",
			Package: pkgName,
		}

		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue
			}

			fieldInfo := p.parseField(field)
			if fieldInfo != nil {
				model.Fields = append(model.Fields, *fieldInfo)
				if fieldInfo.FieldType == "TimeField" {
					model.HasTimeImport = true
				}
			}
		}

		if len(model.Fields) > 0 {
			models = append(models, model)
		}

		return true
	})

	return models
}

// hasDbTag checks if any field in the struct has a db tag
func (p *Parser) hasDbTag(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if field.Tag != nil && strings.Contains(field.Tag.Value, `apolon:"`) {
			return true
		}
	}
	return false
}

// parseField extracts field information from an AST field
func (p *Parser) parseField(field *ast.Field) *FieldInfo {
	if len(field.Names) == 0 {
		return nil
	}

	name := field.Names[0].Name
	goType := p.typeToString(field.Type)
	column := strings.ToLower(name)
	isPK := false

	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")

		for _, part := range strings.Split(tag, " ") {
			if strings.HasPrefix(part, `apolon:"`) {
				value := strings.TrimPrefix(part, `apolon:"`)
				value = strings.TrimSuffix(value, `"`)

				parts := strings.Split(value, ",")
				if len(parts) > 0 && parts[0] != "" {
					column = parts[0]
				}

				for _, opt := range parts[1:] {
					if opt == "pk" {
						isPK = true
					}
				}
			}
		}
	}

	fieldType := p.goTypeToFieldType(goType)
	if fieldType == "" {
		return nil
	}

	return &FieldInfo{
		Name:      name,
		Column:    column,
		FieldType: fieldType,
		GoType:    goType,
		IsPK:      isPK,
	}
}

// typeToString converts an AST type to a string representation
func (p *Parser) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.typeToString(t.Elt)
	}
	return ""
}

// goTypeToFieldType maps Go types to ORM field types
func (p *Parser) goTypeToFieldType(goType string) string {
	switch goType {
	case "int", "int32", "uint", "uint32":
		return "IntField"
	case "int64", "uint64":
		return "Int64Field"
	case "string":
		return "StringField"
	case "bool":
		return "BoolField"
	case "float32", "float64":
		return "FloatField"
	case "time.Time", "*time.Time":
		return "TimeField"
	default:
		return ""
	}
}
