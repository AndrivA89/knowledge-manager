package domain

import (
	"time"
)

type NodeType string

const (
	Concept   NodeType = "CONCEPT"
	Note      NodeType = "NOTE"
	Reference NodeType = "REFERENCE"
)

type Node struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      NodeType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []string  `json:"tags"`
}
