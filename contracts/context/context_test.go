package agentcontext

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestItemTags(t *testing.T) {
	itemType := reflect.TypeOf(Item{})

	expectedTags := map[string]string{
		"Type":     "type",
		"Content":  "content",
		"Source":   "source",
		"Priority": "priority",
		"Tokens":   "tokens",
	}

	for fieldName, expectedJSONTag := range expectedTags {
		field, found := itemType.FieldByName(fieldName)
		if !found {
			t.Errorf("Item missing field %s", fieldName)
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag != expectedJSONTag {
			t.Errorf("field %s: expected json tag %q, got %q", fieldName, expectedJSONTag, jsonTag)
		}
	}
}

func TestApplyBuildOptions(t *testing.T) {
	// Test defaults
	opts := ApplyBuildOptions()
	if opts.MaxTokens != 8192 {
		t.Errorf("expected default MaxTokens 8192, got %d", opts.MaxTokens)
	}
	if len(opts.Sources) != 0 {
		t.Errorf("expected default Sources to be empty, got %v", opts.Sources)
	}
	if opts.Query != "" {
		t.Errorf("expected default Query to be empty, got %q", opts.Query)
	}

	// Test custom options
	opts = ApplyBuildOptions(
		WithMaxTokens(4096),
		WithSources("file", "memory"),
		WithQuery("test query"),
	)

	if opts.MaxTokens != 4096 {
		t.Errorf("expected MaxTokens 4096, got %d", opts.MaxTokens)
	}
	if len(opts.Sources) != 2 || opts.Sources[0] != "file" || opts.Sources[1] != "memory" {
		t.Errorf("expected Sources [file memory], got %v", opts.Sources)
	}
	if opts.Query != "test query" {
		t.Errorf("expected Query %q, got %q", "test query", opts.Query)
	}
}

func TestItemJSON(t *testing.T) {
	item := Item{
		Type:     "file",
		Content:  "content details",
		Source:   "path/to/file",
		Priority: 10,
		Tokens:   100,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal item: %v", err)
	}

	var decoded Item
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal item: %v", err)
	}

	if decoded != item {
		t.Errorf("expected decoded item to match original, got %+v, original %+v", decoded, item)
	}
}
