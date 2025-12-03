package openrouterapigo

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestGenerateSchema_EmptyStruct(t *testing.T) {
	input := struct{}{}
	schema := generateSchema(input)

	if schema["type"] != "object" {
		t.Errorf("Expected type object, got %v", schema["type"])
	}

	if len(schema["properties"].(map[string]interface{})) != 0 {
		t.Errorf("Expected empty properties, got %v", schema["properties"])
	}

	if len(schema["required"].([]string)) != 0 {
		t.Errorf("Expected no required fields, got %v", schema["required"])
	}
}

func TestGenerateSchema_SimpleTypes(t *testing.T) {
	input := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}
	schema := generateSchema(input)

	props := schema["properties"].(map[string]interface{})
	if props["name"].(map[string]interface{})["type"] != "string" {
		t.Error("Expected name type string")
	}

	if props["age"].(map[string]interface{})["type"] != "integer" {
		t.Error("Expected age type integer")
	}

	required := schema["required"].([]string)
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

func TestGenerateSchema_OmitemptyAndDescription(t *testing.T) {
	input := struct {
		ID     string `json:"id,omitempty"`
		Active bool   `json:"active" jsonschema:"Whether user is active"`
	}{}
	schema := generateSchema(input)

	props := schema["properties"].(map[string]interface{})
	if _, ok := props["id"]; !ok {
		t.Error("Expected id property")
	}

	active := props["active"].(map[string]interface{})
	if active["type"] != "boolean" {
		t.Error("Expected active type boolean")
	}

	if active["description"] != "Whether user is active" {
		t.Error("Missing description")
	}

	required := schema["required"].([]string)
	if len(required) != 1 || required[0] != "active" {
		t.Error("Expected only active as required")
	}
}

func TestGenerateSchema_NestedStruct(t *testing.T) {
	input := struct {
		Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		} `json:"address"`
	}{}
	schema := generateSchema(input)

	address := schema["properties"].(map[string]interface{})["address"].(map[string]interface{})
	if address["type"] != "object" {
		t.Error("Expected address type object")
	}

	addressProps := address["properties"].(map[string]interface{})
	if addressProps["street"].(map[string]interface{})["type"] != "string" {
		t.Error("Expected street type string")
	}

	if addressProps["city"].(map[string]interface{})["type"] != "string" {
		t.Error("Expected city type string")
	}
}

func TestGenerateSchema_SliceTypes(t *testing.T) {
	input := struct {
		IDs       []int    `json:"ids"`
		Names     []string `json:"names,omitempty"`
		Addresses []struct {
			City string `json:"city"`
		} `json:"addresses"`
	}{}
	schema := generateSchema(input)

	props := schema["properties"].(map[string]interface{})
	ids := props["ids"].(map[string]interface{})
	if ids["type"] != "array" {
		t.Error("Expected ids type array")
	}
	if ids["items"].(map[string]interface{})["type"] != "integer" {
		t.Error("Expected ids items type integer")
	}

	addresses := props["addresses"].(map[string]interface{})
	if addresses["type"] != "array" {
		t.Error("Expected addresses type array")
	}
	addressItems := addresses["items"].(map[string]interface{})
	if addressItems["type"] != "object" {
		t.Error("Expected addresses items type object")
	}
}

func TestGenerateSchema_NonStructInput(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"String", "test"},
		{"Int", 42},
		{"Slice", []string{"a", "b"}},
		{"Map", map[string]int{"a": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := generateSchema(tt.input)
			schemaStr, _ := json.Marshal(schema)
			if string(schemaStr) != "{}" {
				t.Errorf("Expected empty schema for non-struct input, got %s", schemaStr)
			}
		})
	}
}

func TestToolWrapper_Call(t *testing.T) {
	t.Run("Successfully calls function with correct args", func(t *testing.T) {
		type Args struct {
			A int `json:"a"`
			B int `json:"b"`
		}

		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Function: func(args Args) any {
					return args.A + args.B
				},
			},
		}

		args := json.RawMessage(`{"a": 2, "b": 3}`)
		result, err := wrapper.Call(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if sum, ok := result.(int); !ok || sum != 5 {
			t.Errorf("Expected result 5, got %v", result)
		}
	})

	t.Run("Returns error for invalid JSON", func(t *testing.T) {
		type Args struct {
			Field string `json:"field"`
		}

		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Function: func(args Args) any { return nil },
			},
		}

		_, err := wrapper.Call(json.RawMessage(`{"invalid json`))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("Returns error for wrong argument types", func(t *testing.T) {
		type Args struct {
			Number int `json:"number"`
		}

		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Function: func(args Args) any { return nil },
			},
		}

		_, err := wrapper.Call(json.RawMessage(`{"number": "string"}`))
		if err == nil {
			t.Error("Expected error for type mismatch")
		}
	})
}

func TestToolWrapper_Metadata(t *testing.T) {
	t.Run("Returns correct metadata", func(t *testing.T) {
		type Args struct {
			Name string `json:"name" jsonschema:"description=User name"`
			Age  int    `json:"age"`
		}

		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Name:        "test_tool",
				Description: "Test description",
				Function: func(a Args) any {
					return true
				},
			},
		}

		metadata := wrapper.Metadata()
		if metadata.Name != "test_tool" {
			t.Errorf("Expected name 'test_tool', got '%s'", metadata.Name)
		}

		if metadata.Description != "Test description" {
			t.Errorf("Expected description 'Test description', got '%s'", metadata.Description)
		}
	})
}

func TestToolWrapper_EdgeCases(t *testing.T) {
	t.Run("Handles function returning nil", func(t *testing.T) {
		type Args struct{}
		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Function: func(args Args) any { return nil },
			},
		}

		result, err := wrapper.Call(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("Handles function returning error", func(t *testing.T) {
		type Args struct{}
		expectedErr := errors.New("test error")

		wrapper := &toolWrapper[Args]{
			definition: ToolDefinition[Args]{
				Function: func(args Args) any { return expectedErr },
			},
		}

		result, _ := wrapper.Call(json.RawMessage(`{}`))
		if result != expectedErr {
			t.Errorf("Expected error result, got %v", result)
		}
	})
}

func TestToolWrapper_ReflectionSafety(t *testing.T) {
	t.Run("Panics if not initialized with function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for non-function input")
			}
		}()

		// This should panic
		_ = reflect.ValueOf(42).Call(nil)
	})
}

type fakeTool struct {
	name    string
	params  map[string]interface{}
	callErr error
}

func (f fakeTool) Call(args json.RawMessage) (any, error) {
	return nil, f.callErr
}

func (f fakeTool) Metadata() FunctionDescription {
	return FunctionDescription{
		Name:       f.name,
		Parameters: f.params,
	}
}

func TestToolRegistry_RegisterValidation(t *testing.T) {
	reg := NewToolRegistry()

	if err := reg.Register(fakeTool{name: "", params: map[string]interface{}{}}); err == nil {
		t.Errorf("expected error for empty name")
	}

	if err := reg.Register(fakeTool{name: "dup", params: map[string]interface{}{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := reg.Register(fakeTool{name: "dup", params: map[string]interface{}{}}); err == nil {
		t.Errorf("expected duplicate name error")
	}

	if err := reg.Register(fakeTool{name: "nilparams", params: nil}); err == nil {
		t.Errorf("expected error for nil parameters schema")
	}
}
