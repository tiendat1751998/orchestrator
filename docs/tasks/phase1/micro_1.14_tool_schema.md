# Micro-Task 1.14: Tạo contracts/tool/schema.go

## Thông tin
- **File tạo**: `contracts/tool/schema.go`
- **Package**: `tool`
- **Dependencies trước**: 1.06
- **Thời gian**: 10 phút

## Nội dung CHÍNH XÁC cần tạo

```go
// Package tool defines the contract for tools that AI agents can use.
// Tools are capabilities like reading files, running commands, or searching code.
package tool

// Schema defines the input parameters of a tool using JSON Schema format.
// This is sent to the AI provider so it knows what arguments to pass.
//
// We use JSON Schema because it's the standard format supported by
// OpenAI, Gemini, Claude, and all major AI providers for function calling.
type Schema struct {
	// Type is always "object" for tool parameters.
	Type string `json:"type"`

	// Properties defines each parameter and its type.
	Properties map[string]Property `json:"properties"`

	// Required lists parameter names that must be provided.
	Required []string `json:"required,omitempty"`

	// Description is an overall description of what the parameters represent.
	Description string `json:"description,omitempty"`
}

// Property defines a single parameter in a tool schema.
type Property struct {
	// Type is the JSON type: "string", "integer", "number", "boolean", "array", "object"
	Type string `json:"type"`

	// Description explains what this parameter does.
	// The AI reads this to decide what value to pass.
	// Be specific: "Absolute path to the file" > "The path"
	Description string `json:"description"`

	// Enum restricts the parameter to a set of allowed values.
	// Example: ["json", "yaml", "toml"]
	Enum []string `json:"enum,omitempty"`

	// Default is the default value if the parameter is not provided.
	Default any `json:"default,omitempty"`

	// Items defines the schema for array elements (when Type="array").
	Items *Property `json:"items,omitempty"`

	// Properties defines nested object properties (when Type="object").
	Properties map[string]Property `json:"properties,omitempty"`
}

// NewSchema creates a new Schema with type "object".
func NewSchema() *Schema {
	return &Schema{
		Type:       "object",
		Properties: make(map[string]Property),
	}
}

// AddProperty adds a parameter to the schema.
func (s *Schema) AddProperty(name string, prop Property) *Schema {
	s.Properties[name] = prop
	return s
}

// AddRequired marks a parameter as required.
func (s *Schema) AddRequired(names ...string) *Schema {
	s.Required = append(s.Required, names...)
	return s
}

// StringProperty creates a string-type property.
func StringProperty(description string) Property {
	return Property{Type: "string", Description: description}
}

// IntProperty creates an integer-type property.
func IntProperty(description string) Property {
	return Property{Type: "integer", Description: description}
}

// BoolProperty creates a boolean-type property.
func BoolProperty(description string) Property {
	return Property{Type: "boolean", Description: description}
}

// EnumProperty creates a string property restricted to specific values.
func EnumProperty(description string, values ...string) Property {
	return Property{Type: "string", Description: description, Enum: values}
}
```

## ⚠️ Pitfalls
1. **`Type` phải là `"object"`** cho top-level schema. AI providers reject schemas với type khác.
2. **Property.Items**: Khi Type="array", Items PHẢI có. Nếu thiếu → AI không biết array chứa gì.
3. **Builder pattern**: `AddProperty()` trả về `*Schema` cho chaining: `schema.AddProperty("path", ...).AddRequired("path")`
4. **Helper functions**: `StringProperty()`, `IntProperty()` giảm boilerplate khi build schema.

## Checklist
- [ ] Schema struct với 4 fields
- [ ] Property struct với 6 fields (Type, Description, Enum, Default, Items, Properties)
- [ ] `NewSchema()` constructor
- [ ] Builder methods: `AddProperty()`, `AddRequired()`
- [ ] Helper functions: `StringProperty()`, `IntProperty()`, `BoolProperty()`, `EnumProperty()`
- [ ] `go build ./contracts/...` không lỗi
