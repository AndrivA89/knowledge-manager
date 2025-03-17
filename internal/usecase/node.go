package usecase

import (
	"context"

	"github.com/AndrivA89/neo4j-go-playground/internal/domain"
)

type NodeUseCase struct {
	repo NodeRepository
}

func NewNodeUseCase(repo NodeRepository) *NodeUseCase {
	return &NodeUseCase{
		repo: repo,
	}
}

func (uc *NodeUseCase) CreateNode(ctx context.Context, node *domain.Node) (string, error) {
	return uc.repo.CreateNode(ctx, node)
}

func (uc *NodeUseCase) GetNode(ctx context.Context, id string) (*domain.Node, error) {
	return uc.repo.GetNodeByID(ctx, id)
}

func (uc *NodeUseCase) CreateRelationship(ctx context.Context, rel *domain.Relationship) ([]string, error) {
	return uc.repo.CreateRelationship(ctx, rel)
}

func (uc *NodeUseCase) UpdateNode(ctx context.Context, node *domain.Node) error {
	return uc.repo.UpdateNode(ctx, node)
}

func (uc *NodeUseCase) DeleteNode(ctx context.Context, id string) error {
	return uc.repo.DeleteNode(ctx, id)
}

func (uc *NodeUseCase) DeleteRelationship(ctx context.Context, relationshipID string) error {
	return uc.repo.DeleteRelationship(ctx, relationshipID)
}

func (uc *NodeUseCase) SearchNodes(ctx context.Context, query, criteria string) ([]*domain.Node, error) {
	return uc.repo.SearchNodes(ctx, query, criteria)
}
