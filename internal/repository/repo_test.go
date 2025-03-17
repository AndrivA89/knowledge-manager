package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/AndrivA89/neo4j-go-playground/internal/domain"
)

var testDriver neo4j.DriverWithContext

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not connect to docker: %s\n", err)
		os.Exit(1)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "neo4j",
		Tag:        "4.4",
		Env: []string{
			"NEO4J_AUTH=neo4j/password",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
	})
	if err != nil {
		fmt.Printf("Could not start resource: %s\n", err)
		os.Exit(1)
	}

	pool.MaxWait = 120 * time.Second

	if err := pool.Retry(func() error {
		var err error
		testDriver, err = neo4j.NewDriverWithContext(
			"bolt://localhost:"+resource.GetPort("7687/tcp"),
			neo4j.BasicAuth("neo4j", "password", ""),
		)
		if err != nil {
			return err
		}
		return testDriver.VerifyConnectivity(context.Background())
	}); err != nil {
		fmt.Printf("Could not connect to docker: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		fmt.Printf("Could not purge resource: %s\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func TestCreateUpdateDeleteNode(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	node := &domain.Node{
		Title:   "Test Node",
		Content: "Test content",
		Type:    domain.Concept,
		Tags:    []string{"test", "node"},
	}
	id, err := repo.CreateNode(ctx, node)
	assert.NoError(t, err, "CreateNode error should be nil")
	assert.NotEqual(t, "", id, "Node id should not be empty")
	t.Logf("Created node with id: %s", id)

	created, err := repo.GetNodeByID(ctx, id)
	assert.NoError(t, err, "GetNodeByID error should be nil")
	assert.Equal(t, node.Title, created.Title, "Titles should match")

	created.Title = "Updated Test Node"
	created.Content = "Updated content"
	err = repo.UpdateNode(ctx, created)
	assert.NoError(t, err, "UpdateNode error should be nil")

	updated, err := repo.GetNodeByID(ctx, id)
	assert.NoError(t, err, "GetNodeByID after update error should be nil")
	assert.Equal(t, "Updated Test Node", updated.Title, "Title should be updated")

	err = repo.DeleteNode(ctx, id)
	assert.NoError(t, err, "DeleteNode error should be nil")

	_, err = repo.GetNodeByID(ctx, id)
	assert.Error(t, err, "GetNodeByID should return error for deleted node")
}

func TestCreateDeleteRelationship(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	node1 := &domain.Node{
		Title:   "Rel Node 1",
		Content: "Content 1",
		Type:    domain.Concept,
		Tags:    []string{"tag1"},
	}
	node2 := &domain.Node{
		Title:   "Rel Node 2",
		Content: "Content 2",
		Type:    domain.Concept,
		Tags:    []string{"tag2"},
	}

	id1, err := repo.CreateNode(ctx, node1)
	assert.NoError(t, err, "CreateNode for node1 should succeed")
	id2, err := repo.CreateNode(ctx, node2)
	assert.NoError(t, err, "CreateNode for node2 should succeed")
	node1.ID = id1
	node2.ID = id2

	time.Sleep(500 * time.Millisecond)

	rel := &domain.Relationship{
		SourceID:    node1.ID,
		TargetIDs:   []string{node2.ID},
		Type:        domain.RelatedTo,
		Description: "Test relationship",
	}
	relIDs, err := repo.CreateRelationship(ctx, rel)
	assert.NoError(t, err, "CreateRelationship error should be nil")
	assert.Len(t, relIDs, 1, "Should have 1 relationship id")
	relID := relIDs[0]
	t.Logf("Created relationship with id: %s", relID)

	err = repo.DeleteRelationship(ctx, relID)
	assert.NoError(t, err, "DeleteRelationship error should be nil")

	err = repo.DeleteNode(ctx, node1.ID)
	assert.NoError(t, err, "DeleteNode for node1 should succeed")
	err = repo.DeleteNode(ctx, node2.ID)
	assert.NoError(t, err, "DeleteNode for node2 should succeed")
}

func TestMultipleRelationships(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	source := &domain.Node{
		Title:   "Source Node",
		Content: "Source Content",
		Type:    domain.Concept,
		Tags:    []string{"source"},
	}
	target1 := &domain.Node{
		Title:   "Target Node 1",
		Content: "Target Content 1",
		Type:    domain.Concept,
		Tags:    []string{"target1"},
	}
	target2 := &domain.Node{
		Title:   "Target Node 2",
		Content: "Target Content 2",
		Type:    domain.Concept,
		Tags:    []string{"target2"},
	}

	sid, err := repo.CreateNode(ctx, source)
	assert.NoError(t, err, "CreateNode for source should succeed")
	tid1, err := repo.CreateNode(ctx, target1)
	assert.NoError(t, err, "CreateNode for target1 should succeed")
	tid2, err := repo.CreateNode(ctx, target2)
	assert.NoError(t, err, "CreateNode for target2 should succeed")
	source.ID = sid
	target1.ID = tid1
	target2.ID = tid2

	rel := &domain.Relationship{
		SourceID:    source.ID,
		TargetIDs:   []string{target1.ID, target2.ID},
		Type:        domain.RelatedTo,
		Description: "Multiple relationship test",
	}
	relIDs, err := repo.CreateRelationship(ctx, rel)
	assert.NoError(t, err, "CreateRelationship error should be nil")
	assert.Len(t, relIDs, 2, "Should have 2 relationship ids")

	for _, rid := range relIDs {
		err = repo.DeleteRelationship(ctx, rid)
		assert.NoError(t, err, "DeleteRelationship error should be nil for id %s", rid)
	}

	err = repo.DeleteNode(ctx, source.ID)
	assert.NoError(t, err, "DeleteNode for source should succeed")
	err = repo.DeleteNode(ctx, target1.ID)
	assert.NoError(t, err, "DeleteNode for target1 should succeed")
	err = repo.DeleteNode(ctx, target2.ID)
	assert.NoError(t, err, "DeleteNode for target2 should succeed")
}

func TestSearchNodesByTag(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	// Create two nodes.
	node1 := &domain.Node{
		Title:   "Golang Tutorial",
		Content: "Learn how to use Go with Neo4j",
		Type:    domain.Concept,
		Tags:    []string{"golang", "neo4j", "tutorial"},
	}
	node2 := &domain.Node{
		Title:   "Graph Databases",
		Content: "Overview of graph databases",
		Type:    domain.Concept,
		Tags:    []string{"database", "graph"},
	}
	id1, err := repo.CreateNode(ctx, node1)
	assert.NoError(t, err, "CreateNode for node1 should succeed")
	id2, err := repo.CreateNode(ctx, node2)
	assert.NoError(t, err, "CreateNode for node2 should succeed")
	node1.ID = id1
	node2.ID = id2

	// Allow some time for Neo4j to process.
	time.Sleep(500 * time.Millisecond)

	// Search by tag "neo4j" using criteria "Tag".
	results, err := repo.SearchNodes(ctx, "neo4j", "Tag")
	assert.NoError(t, err, "SearchNodes (Tag) should not error")
	// Expect only node1 to be found.
	assert.Len(t, results, 1, "Should find one node for tag 'neo4j'")
	assert.Equal(t, node1.Title, results[0].Title, "Found node should match node1")

	// Cleanup.
	err = repo.DeleteNode(ctx, node1.ID)
	assert.NoError(t, err, "DeleteNode for node1 should succeed")
	err = repo.DeleteNode(ctx, node2.ID)
	assert.NoError(t, err, "DeleteNode for node2 should succeed")
}

func TestSearchNodesByTitleContent(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	node1 := &domain.Node{
		Title:   "Learning Golang",
		Content: "This is a tutorial for Golang.",
		Type:    domain.Concept,
		Tags:    []string{"golang"},
	}
	node2 := &domain.Node{
		Title:   "Graph Databases",
		Content: "Neo4j is a popular graph database.",
		Type:    domain.Concept,
		Tags:    []string{"neo4j", "graph"},
	}

	id1, err := repo.CreateNode(ctx, node1)
	assert.NoError(t, err, "CreateNode for node1 should succeed")
	id2, err := repo.CreateNode(ctx, node2)
	assert.NoError(t, err, "CreateNode for node2 should succeed")
	node1.ID = id1
	node2.ID = id2

	time.Sleep(500 * time.Millisecond)

	// Search by title/content "Golang".
	results, err := repo.SearchNodes(ctx, "Golang", "Title/Content")
	assert.NoError(t, err, "SearchNodes (Title/Content) should not error")
	assert.Len(t, results, 1, "Should find one node containing 'Golang'")
	assert.Equal(t, node1.Title, results[0].Title, "Found node should match node1")

	// Cleanup.
	err = repo.DeleteNode(ctx, node1.ID)
	assert.NoError(t, err, "DeleteNode for node1 should succeed")
	err = repo.DeleteNode(ctx, node2.ID)
	assert.NoError(t, err, "DeleteNode for node2 should succeed")
}

func TestSearchNodesAll(t *testing.T) {
	ctx := context.Background()
	repo := NewNodeRepository(testDriver)

	node1 := &domain.Node{
		Title:   "Learn Golang",
		Content: "Golang tutorial and tips",
		Type:    domain.Concept,
		Tags:    []string{"golang", "tutorial"},
	}
	node2 := &domain.Node{
		Title:   "Graph Theory",
		Content: "An introduction to graph theory",
		Type:    domain.Concept,
		Tags:    []string{"graph", "math"},
	}

	id1, err := repo.CreateNode(ctx, node1)
	assert.NoError(t, err, "CreateNode for node1 should succeed")
	id2, err := repo.CreateNode(ctx, node2)
	assert.NoError(t, err, "CreateNode for node2 should succeed")
	node1.ID = id1
	node2.ID = id2

	time.Sleep(500 * time.Millisecond)

	// Search with query "graph" using criteria "All"
	results, err := repo.SearchNodes(ctx, "graph", "All")
	assert.NoError(t, err, "SearchNodes (All) should not error")
	// Expect to find node2 (title contains "Graph" and tag "graph").
	assert.Len(t, results, 1, "Should find one node for query 'graph'")
	assert.Equal(t, node2.Title, results[0].Title, "Found node should match node2")

	// Cleanup.
	err = repo.DeleteNode(ctx, node1.ID)
	assert.NoError(t, err, "DeleteNode for node1 should succeed")
	err = repo.DeleteNode(ctx, node2.ID)
	assert.NoError(t, err, "DeleteNode for node2 should succeed")
}
