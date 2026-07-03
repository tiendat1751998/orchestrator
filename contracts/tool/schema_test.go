package tool

import (
	"testing"
)

func TestSchemaBuilder(t *testing.T) {
	s := NewSchema().
		AddProperty("name", StringProperty("The name of the resource")).
		AddProperty("count", IntProperty("The number of items")).
		AddProperty("enabled", BoolProperty("Whether it is enabled")).
		AddProperty("role", EnumProperty("The role", "admin", "user")).
		AddRequired("name", "count")

	if s.Type != "object" {
		t.Errorf("expected Type to be 'object', got %q", s.Type)
	}

	if len(s.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(s.Properties))
	}

	if nameProp, ok := s.Properties["name"]; !ok || nameProp.Type != "string" {
		t.Errorf("expected 'name' to be string property, got %+v", nameProp)
	}

	if len(s.Required) != 2 || s.Required[0] != "name" || s.Required[1] != "count" {
		t.Errorf("expected required fields [name, count], got %v", s.Required)
	}

	if s.Raw() != s {
		t.Errorf("expected Raw() to return the schema itself")
	}
}
