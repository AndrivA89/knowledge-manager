# Neo4j Go Playground

**Neo4j Go Playground** is a sample project for learning and experimenting with Neo4j using the Go programming language and the Fyne GUI toolkit. This project demonstrates basic CRUD operations with a graph database, along with a simple interactive UI for visualizing and managing nodes and relationships.

## Features

- **Graph Visualization**: Display nodes and relationships on a canvas.
- **CRUD Operations**: Create, update, and delete nodes and relationships in Neo4j.
- **Search Functionality**: Search nodes by tags, title, or content.
- **Interactive UI**: Edit and delete nodes directly by clicking on them.
- **Relationship Management**: Add and remove relationships between nodes.
- **Modular Code**: Clean and refactored code structure for easy learning.

## Technologies

- **Go**: Main programming language.
- **Neo4j**: Graph database.
- **Fyne**: Cross-platform GUI toolkit for Go.
- **Dockertest**: For integration testing with Docker.
- **Testify**: For writing expressive tests.

## Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/AndrivA89/neo4j-go-playground.git
   cd neo4j-go-playground
   
2. **Install Dependencies**:

   Make sure you have Go installed (version 1.16+ recommended).
   ```bash
   go mod download
   ```
   
3. **Set up Neo4j**:

   You can run Neo4j using Docker. For example:
   ```bash
   docker run --name neo4j -p7474:7474 -p7687:7687 -e NEO4J_AUTH=neo4j/password neo4j:4.4
   ```
   or run docker-compose file in current project:
   ```bash
   docker-compose up -d
   ```

4. **Run the Application**:

   ```bash
   go run ./cmd/app
   ```

5. Testing

   This project includes integration tests using ory/dockertest and testify/assert. To run tests with Docker:
   ```bash
   go test -tags=docker_test ./internal/repository -v
   ```
   Ensure Docker is running on your machine.


