package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TypeGenConfig holds configuration for type generation
type TypeGenConfig struct {
	OutputDir    string   `json:"outputDir"`
	Packages     []string `json:"packages"`
	IgnoreFields []string `json:"ignoreFields"`
	CustomTypes  map[string]string `json:"customTypes"`
}

// TypeInfo represents a Go type that will be converted to TypeScript
type TypeInfo struct {
	Name        string            `json:"name"`
	Fields      []FieldInfo       `json:"fields"`
	Package     string            `json:"package"`
	Comments    []string          `json:"comments"`
	Annotations map[string]string `json:"annotations"`
}

// FieldInfo represents a field in a Go struct
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	JSONTag  string `json:"jsonTag"`
	Optional bool   `json:"optional"`
	Comment  string `json:"comment"`
}

var config = TypeGenConfig{
	OutputDir: "../../packages/shared-types/api",
	Packages:  []string{"./internal/models", "./internal/handlers", "./internal/types"},
	IgnoreFields: []string{"CreatedAt", "UpdatedAt", "DeletedAt"},
	CustomTypes: map[string]string{
		"time.Time":     "string", // ISO date strings
		"uuid.UUID":     "string",
		"json.RawMessage": "any",
		"firestore.DocumentRef": "string", // Document path
	},
}

func main() {
	if err := generateTypes(); err != nil {
		fmt.Printf("Error generating types: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ TypeScript interfaces generated successfully!")
}

func generateTypes() error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	var allTypes []TypeInfo
	
	// Parse each package
	for _, pkg := range config.Packages {
		types, err := parsePackage(pkg)
		if err != nil {
			return fmt.Errorf("failed to parse package %s: %v", pkg, err)
		}
		allTypes = append(allTypes, types...)
	}

	// Group types by category for better organization
	categories := map[string][]TypeInfo{
		"models":    {},
		"requests":  {},
		"responses": {},
		"common":    {},
	}

	for _, t := range allTypes {
		category := categorizeType(t)
		categories[category] = append(categories[category], t)
	}

	// Generate TypeScript files
	for category, types := range categories {
		if len(types) == 0 {
			continue
		}
		
		if err := generateTypeScriptFile(category, types); err != nil {
			return fmt.Errorf("failed to generate %s types: %v", category, err)
		}
	}

	// Generate index file
	if err := generateIndexFile(categories); err != nil {
		return fmt.Errorf("failed to generate index file: %v", err)
	}

	return nil
}

func parsePackage(packagePath string) ([]TypeInfo, error) {
	var types []TypeInfo
	
	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileTypes, err := parseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %v", path, err)
		}
		
		types = append(types, fileTypes...)
		return nil
	})

	return types, err
}

func parseFile(filename string) ([]TypeInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var types []TypeInfo
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				typeInfo := TypeInfo{
					Name:        x.Name.Name,
					Package:     node.Name.Name,
					Comments:    extractComments(x.Doc),
					Annotations: make(map[string]string),
					Fields:      []FieldInfo{},
				}

				// Parse struct fields
				for _, field := range structType.Fields.List {
					fieldInfo := parseField(field)
					if shouldIncludeField(fieldInfo) {
						typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
					}
				}

				// Only include structs with exported fields
				if len(typeInfo.Fields) > 0 && isExported(typeInfo.Name) {
					types = append(types, typeInfo)
				}
			}
		}
		return true
	})

	return types, nil
}

func parseField(field *ast.Field) FieldInfo {
	fieldInfo := FieldInfo{}
	
	if len(field.Names) > 0 {
		fieldInfo.Name = field.Names[0].Name
	}

	fieldInfo.Type = typeToString(field.Type)
	fieldInfo.Comment = extractFieldComment(field.Doc)

	// Parse struct tags
	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")
		if jsonTag := extractJSONTag(tag); jsonTag != "" {
			fieldInfo.JSONTag = jsonTag
			fieldInfo.Optional = strings.Contains(tag, "omitempty")
		}
	}

	return fieldInfo
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", typeToString(t.X), t.Sel.Name)
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", typeToString(t.Elt))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", typeToString(t.X))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeToString(t.Key), typeToString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

func goTypeToTypeScript(goType string) string {
	// Handle custom type mappings
	if tsType, exists := config.CustomTypes[goType]; exists {
		return tsType
	}

	// Handle basic Go types
	switch goType {
	case "string", "[]byte":
		return "string"
	case "int", "int8", "int16", "int32", "int64", 
		 "uint", "uint8", "uint16", "uint32", "uint64",
		 "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "interface{}", "any":
		return "any"
	}

	// Handle slices
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		return fmt.Sprintf("%s[]", goTypeToTypeScript(elemType))
	}

	// Handle pointers
	if strings.HasPrefix(goType, "*") {
		elemType := strings.TrimPrefix(goType, "*")
		return fmt.Sprintf("%s | null", goTypeToTypeScript(elemType))
	}

	// Handle maps
	if strings.HasPrefix(goType, "map[") {
		return "{ [key: string]: any }"
	}

	// Default to the type name (assume it's a custom type)
	return goType
}

func generateTypeScriptFile(category string, types []TypeInfo) error {
	filename := filepath.Join(config.OutputDir, fmt.Sprintf("%s.ts", category))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "// Generated by typegen at %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "// DO NOT EDIT - This file is automatically generated\n\n")

	// Write types
	for _, t := range types {
		if err := writeTypeScript(file, t); err != nil {
			return err
		}
	}

	return nil
}

func writeTypeScript(file *os.File, t TypeInfo) error {
	// Write comments
	for _, comment := range t.Comments {
		fmt.Fprintf(file, "// %s\n", comment)
	}

	fmt.Fprintf(file, "export interface %s {\n", t.Name)
	
	for _, field := range t.Fields {
		jsonName := field.JSONTag
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name[:1]) + field.Name[1:]
		}
		
		optional := ""
		if field.Optional {
			optional = "?"
		}

		if field.Comment != "" {
			fmt.Fprintf(file, "  // %s\n", field.Comment)
		}
		
		fmt.Fprintf(file, "  %s%s: %s;\n", jsonName, optional, goTypeToTypeScript(field.Type))
	}
	
	fmt.Fprintf(file, "}\n\n")
	return nil
}

func generateIndexFile(categories map[string][]TypeInfo) error {
	filename := filepath.Join(config.OutputDir, "index.ts")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "// Generated by typegen at %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "// DO NOT EDIT - This file is automatically generated\n\n")

	for category := range categories {
		fmt.Fprintf(file, "export * from './%s';\n", category)
	}

	return nil
}

func categorizeType(t TypeInfo) string {
	name := strings.ToLower(t.Name)
	
	if strings.Contains(name, "request") || strings.HasSuffix(name, "req") {
		return "requests"
	}
	if strings.Contains(name, "response") || strings.HasSuffix(name, "resp") {
		return "responses"
	}
	if t.Package == "models" {
		return "models"
	}
	return "common"
}

func shouldIncludeField(field FieldInfo) bool {
	if !isExported(field.Name) {
		return false
	}
	
	for _, ignore := range config.IgnoreFields {
		if field.Name == ignore {
			return false
		}
	}
	
	return true
}

func isExported(name string) bool {
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}

func extractComments(doc *ast.CommentGroup) []string {
	if doc == nil {
		return nil
	}
	
	var comments []string
	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			comments = append(comments, text)
		}
	}
	return comments
}

func extractFieldComment(doc *ast.CommentGroup) string {
	comments := extractComments(doc)
	if len(comments) > 0 {
		return comments[0]
	}
	return ""
}

func extractJSONTag(tag string) string {
	// Simple JSON tag extraction
	if strings.Contains(tag, "json:") {
		start := strings.Index(tag, "json:\"") + 6
		end := strings.Index(tag[start:], "\"")
		if end > 0 {
			jsonTag := tag[start : start+end]
			if comma := strings.Index(jsonTag, ","); comma > 0 {
				return jsonTag[:comma]
			}
			return jsonTag
		}
	}
	return ""
}