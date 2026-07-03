// Package knowledge defines contracts for the local SQLite knowledge store, skill graph, and product memory.
package knowledge

import (
	"context"
)

// EntryType categorizes knowledge entries.
type EntryType string

const (
	EntryPattern  EntryType = "pattern"
	EntryTemplate EntryType = "template"
	EntryFact     EntryType = "fact"
	EntryDecision EntryType = "decision"
)

// Entry is a single unit of knowledge.
type Entry struct {
	ID           string    `json:"id"`
	Type         EntryType `json:"type"`
	Title        string    `json:"title"`
	Content      any       `json:"content"`
	Tags         []string  `json:"tags,omitempty"`
	Score        float64   `json:"score"`
	UsedCount    int       `json:"used_count"`
	SuccessCount int       `json:"success_count"`
}

// KnowledgeStore provides core persistence operations for knowledge units.
type KnowledgeStore interface {
	Save(ctx context.Context, entry Entry) error
	Get(ctx context.Context, id string) (*Entry, error)
	Delete(ctx context.Context, id string) error
}

// SkillNode represents a specific technical competency.
type SkillNode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// SkillEdge represents dependency relationship.
type SkillEdge struct {
	FromID   string `json:"from_id"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"`
}

// SkillGraph provides graph traversal queries over technology nodes (RFC-0032).
type SkillGraph interface {
	AddSkill(ctx context.Context, skill SkillNode) error
	LinkSkills(ctx context.Context, edge SkillEdge) error
	ResolveDependencies(ctx context.Context, skillID string) ([]SkillNode, error)
}

// ProductPattern represents a business feature layout.
type ProductPattern struct {
	ID          string   `json:"id"`
	FeatureName string   `json:"feature_name"`
	Domain      string   `json:"domain"`
	Description string   `json:"description"`
	Components  []string `json:"components"`
}

// ProductMemory provides access to business patterns (RFC-0041).
type ProductMemory interface {
	AddPattern(ctx context.Context, pattern ProductPattern) error
	MatchPattern(ctx context.Context, goalQuery string) ([]ProductPattern, error)
}

// DecayParams configures the depreciation rate.
type DecayParams struct {
	DecayFactor   float64 `json:"decay_factor"`
	MinConfidence float64 `json:"min_confidence"`
}

// KnowledgeDecayer deprecates stale or incorrect nodes (RFC-0051).
type KnowledgeDecayer interface {
	DeprecateNode(ctx context.Context, nodeID string, params DecayParams) error
	PruneStaleNodes(ctx context.Context, minConfidence float64) (int, error)
}
