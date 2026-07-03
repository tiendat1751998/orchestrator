# Micro-Task 1.14: Create contracts/tool/schema.go

## Info
- **File**: `contracts/tool/schema.go`
- **Package**: `tool`
- **Depends on**: 1.06
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Defines the input parameters schema of tools using the JSON Schema format.

## EXACT code to create

```go
// Package tool defines the contracts for tools that AI agents can use.
package tool

// Schema defines the input parameters of a tool using JSON Schema format.
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
	// Type is the JSON type ("string", "integer", "number", "boolean", "array", "object").
	Type string `json:"type"`

	// Description explains what this parameter does.
	Description string `json:"description"`

	// Enum restricts the parameter to a set of allowed values.
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

// Raw returns the underlying Schema object itself.
// This is used to conform to schemas mapping across providers.
func (s *Schema) Raw() any {
	return s
}
```

## ⚠️ Pitfalls

### Pitfall 1: Top-level schema Type other than "object"
```go
schema.Type = "object" // Top-level is always structured properties map.
```
Ensure `NewSchema()` initializes `Type` as `"object"`.

### Pitfall 2: Neglecting the Items field for Array types
If you define `Type: "array"` but omit the `Items` definition pointer, the AI will not know what types the array contains (e.g. array of strings, array of integers), causing execution failures.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File `contracts/tool/schema.go` exists
- [ ] Package: `tool`
- [ ] `Schema` and `Property` structs exist with correct JSON tags
- [ ] `NewSchema()` initializes `Type` as `"object"`
- [ ] Builder methods (`AddProperty`, `AddRequired`) support method chaining
- [ ] Helpers like `StringProperty`, `IntProperty`, `BoolProperty` exist
- [ ] `Raw()` method returns the schema representation
- [ ] `go build ./contracts/...` passes
