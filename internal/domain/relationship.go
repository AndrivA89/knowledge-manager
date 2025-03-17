package domain

import "time"

type RelationType string

const (
	RelatedTo    RelationType = "RELATED_TO"
	References   RelationType = "REFERENCES"
	IsPartOf     RelationType = "IS_PART_OF"
	HasPart      RelationType = "HAS_PART"
	DependsOn    RelationType = "DEPENDS_ON"
	IsPrecededBy RelationType = "IS_PRECEDED_BY"
)

type Relationship struct {
	ID          string       `json:"id"`
	SourceID    string       `json:"source_id"`
	TargetIDs   []string     `json:"target_ids"`
	Type        RelationType `json:"type"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
}
