package openrouterapigo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ToolDefinition[T any] struct {
	Function    func(T) any
	Name        string
	Description string
}

type ToolMetadata struct {
	Name        string
	Description string
	ArgsSchema  string // JSON Schema
}

type ToolInterface interface {
	Call(args json.RawMessage) (any, error)
	Metadata() FunctionDescription
}

type toolWrapper[T any] struct {
	definition ToolDefinition[T]
}

func (tw toolWrapper[T]) Call(args json.RawMessage) (any, error) {
	var input T
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	return tw.definition.Function(input), nil
}

func (tw toolWrapper[T]) Metadata() FunctionDescription {
	schema := generateSchema(reflect.New(reflect.TypeOf(tw.definition.Function).In(0)).Elem().Interface())
	return FunctionDescription{
		Description: tw.definition.Description,
		Name:        tw.definition.Name,
		Parameters:  schema,
	}
}

func generateSchema(input any) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}

	t := reflect.TypeOf(input)
	if t.Kind() != reflect.Struct {
		return map[string]interface{}{}
	}

	// Process struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract JSON field name
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		// Create property schema
		propSchema := map[string]interface{}{}
		fieldType := field.Type

		// Handle different types
		switch fieldType.Kind() {
		case reflect.String:
			propSchema["type"] = "string"

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			propSchema["type"] = "integer"

		case reflect.Float32, reflect.Float64:
			propSchema["type"] = "number"

		case reflect.Bool:
			propSchema["type"] = "boolean"

		case reflect.Struct:
			// Recursive call for nested structs
			nested := reflect.New(fieldType).Elem().Interface()
			propSchema = generateSchema(nested)

		case reflect.Slice, reflect.Array:
			propSchema["type"] = "array"
			elemType := fieldType.Elem()

			// Handle slice of structs
			if elemType.Kind() == reflect.Struct {
				nested := reflect.New(elemType).Elem().Interface()
				propSchema["items"] = generateSchema(nested)
			} else {
				// Handle basic types in slices
				propSchema["items"] = map[string]interface{}{
					"type": goTypeToJSONType(elemType.Kind()),
				}
			}

		default:
			propSchema["type"] = "object"
		}

		// Add description from struct tag if available
		if desc := field.Tag.Get("jsonschema"); desc != "" {
			propSchema["description"] = desc
		} else if desc := field.Tag.Get("desc"); desc != "" {
			propSchema["description"] = desc
		}

		// Add to properties
		schema["properties"].(map[string]interface{})[fieldName] = propSchema

		// Check if required
		if !strings.Contains(jsonTag, "omitempty") {
			schema["required"] = append(schema["required"].([]string), fieldName)
		}
	}

	return schema
}

func goTypeToJSONType(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return "object"
	}
}

type ToolRegistry struct {
	tools map[string]ToolInterface
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolInterface),
	}
}

func (r *ToolRegistry) Register(tool ToolInterface) {
	r.tools[tool.Metadata().Name] = tool
}

func (r *ToolRegistry) GenerateTools() ([]Tool, error) {
	metadata := make([]Tool, len(r.tools))
	i := 0
	for _, tool := range r.tools {
		metadata[i] = Tool{
			Type:     DefaultToolType,
			Function: tool.Metadata(),
		}
		i++
	}
	return metadata, nil
}

func (r *ToolRegistry) CallTool(name string, args json.RawMessage) (string, error) {
	for _, tool := range r.tools {
		if tool.Metadata().Name == name {
			returnedValue, err := tool.Call(args)
			if err != nil {
				return "", err
			}
			jsonData, err := json.Marshal(returnedValue)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		}
	}
	return "", fmt.Errorf("tool not found: %s", name)
}
