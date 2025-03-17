package usecase

import (
	"context"

	"github.com/AndrivA89/neo4j-go-playground/internal/domain"
)

type NodeRepository interface {
	CreateNode(context.Context, *domain.Node) (string, error)
	GetNodeByID(context.Context, string) (*domain.Node, error)
	CreateRelationship(ctx context.Context, rel *domain.Relationship) ([]string, error)
	UpdateNode(ctx context.Context, node *domain.Node) error
	DeleteNode(ctx context.Context, id string) error
	DeleteRelationship(ctx context.Context, relationshipID string) error
}
