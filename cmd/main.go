package main

import (
	"context"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/AndrivA89/neo4j-go-playground/internal/domain"
	"github.com/AndrivA89/neo4j-go-playground/internal/repository"
	"github.com/AndrivA89/neo4j-go-playground/internal/ui"
	"github.com/AndrivA89/neo4j-go-playground/internal/usecase"
)

func main() {
	neo4jUri := "bolt://localhost:7687"
	neo4jUsername := "neo4j"
	neo4jPassword := "password"

	driver, err := neo4j.NewDriverWithContext(neo4jUri, neo4j.BasicAuth(neo4jUsername, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer func() {
		if err = driver.Close(context.Background()); err != nil {
			log.Printf("Error closing Neo4j driver: %v", err)
		}
	}()

	repo := repository.NewNodeRepository(driver)
	nodeUseCase := usecase.NewNodeUseCase(repo)

	sampleNode1 := &domain.Node{
		ID:      "1",
		Title:   "First Node",
		Content: "Sample Note",
		Type:    domain.Note,
		Tags:    []string{"1", "2"},
	}

	sampleNode2 := &domain.Node{
		ID:      "2",
		Title:   "Second Node",
		Content: "Reference",
		Type:    domain.Reference,
		Tags:    []string{"3", "4"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := nodeUseCase.CreateNode(ctx, sampleNode1)
	if err != nil {
		log.Fatalf("Failed to create first node: %v", err)
	}
	sampleNode1.ID = id

	id, err = nodeUseCase.CreateNode(ctx, sampleNode2)
	if err != nil {
		log.Fatalf("Failed to create second node: %v", err)
	}
	sampleNode2.ID = id

	nodes := []*domain.Node{sampleNode1, sampleNode2}

	ui.ShowGraphUI(nodeUseCase, nodes, []ui.Edge{})
}
