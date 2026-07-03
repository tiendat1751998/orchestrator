package agent

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestManifestTags(t *testing.T) {
	mType := reflect.TypeOf(Manifest{})

	expectedTags := map[string]struct {
		yaml string
		json string
	}{
		"Name":         {yaml: "name", json: "name"},
		"Version":      {yaml: "version", json: "version"},
		"Role":         {yaml: "role", json: "role"},
		"Description":  {yaml: "description", json: "description"},
		"Capabilities": {yaml: "capabilities", json: "capabilities"},
		"Provider":     {yaml: "provider", json: "provider"},
		"Model":        {yaml: "model,omitempty", json: "model,omitempty"},
		"Tools":        {yaml: "tools,omitempty", json: "tools,omitempty"},
		"SystemPrompt": {yaml: "system_prompt,omitempty", json: "system_prompt,omitempty"},
		"PromptFile":   {yaml: "prompt_file,omitempty", json: "prompt_file,omitempty"},
		"Temperature":  {yaml: "temperature,omitempty", json: "temperature,omitempty"},
		"MaxTokens":    {yaml: "max_tokens,omitempty", json: "max_tokens,omitempty"},
	}

	for fieldName, tags := range expectedTags {
		field, found := mType.FieldByName(fieldName)
		if !found {
			t.Errorf("Manifest missing field %s", fieldName)
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag != tags.yaml {
			t.Errorf("field %s: expected yaml tag %q, got %q", fieldName, tags.yaml, yamlTag)
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag != tags.json {
			t.Errorf("field %s: expected json tag %q, got %q", fieldName, tags.json, jsonTag)
		}
	}
}

func TestManifestJSON(t *testing.T) {
	manifest := Manifest{
		Name:         "test-agent",
		Version:      "1.0.0",
		Role:         "tester",
		Description:  "testing agent manifest",
		Capabilities: []Capability{CapabilityTesting},
		Provider:     "mock-provider",
		Temperature:  0.7,
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	jsonStr := string(data)

	// Check required/non-empty fields are present
	if !strings.Contains(jsonStr, `"name":"test-agent"`) {
		t.Errorf("expected json to contain name tag, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"capabilities":["testing"]`) {
		t.Errorf("expected json to contain capabilities, got: %s", jsonStr)
	}

	// Check optional/empty fields are omitted due to omitempty
	if strings.Contains(jsonStr, `"model"`) {
		t.Errorf("expected json to omit empty model, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"tools"`) {
		t.Errorf("expected json to omit empty tools, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"system_prompt"`) {
		t.Errorf("expected json to omit empty system_prompt, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"prompt_file"`) {
		t.Errorf("expected json to omit empty prompt_file, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"max_tokens"`) {
		t.Errorf("expected json to omit empty max_tokens, got: %s", jsonStr)
	}

	// Now check deserialization
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}

	if decoded.Name != manifest.Name {
		t.Errorf("expected Name %q, got %q", manifest.Name, decoded.Name)
	}
	if len(decoded.Capabilities) != 1 || decoded.Capabilities[0] != CapabilityTesting {
		t.Errorf("expected capabilities [testing], got %v", decoded.Capabilities)
	}
	if decoded.Temperature != 0.7 {
		t.Errorf("expected Temperature 0.7, got %f", decoded.Temperature)
	}
}
