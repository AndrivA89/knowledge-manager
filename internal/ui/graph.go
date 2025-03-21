package ui

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/AndrivA89/neo4j-go-playground/internal/domain"
	"github.com/AndrivA89/neo4j-go-playground/internal/usecase"
)

// Edge represents a relationship between two nodes.
type Edge struct {
	ID   string
	From *domain.Node
	To   *domain.Node
	Type string
}

// NodeWidget is a custom widget to display a Node along with edit functionality.
type NodeWidget struct {
	widget.BaseWidget
	Node         *domain.Node
	Pos          fyne.Position
	ParentWindow fyne.Window
	UseCase      *usecase.NodeUseCase
	OnDelete     func(*domain.Node)
	OnUpdate     func(*domain.Node)
}

// NewNodeWidget creates a new NodeWidget.
func NewNodeWidget(n *domain.Node, pos fyne.Position, w fyne.Window, uc *usecase.NodeUseCase, onDelete, onUpdate func(*domain.Node)) *NodeWidget {
	nw := &NodeWidget{
		Node:         n,
		Pos:          pos,
		ParentWindow: w,
		UseCase:      uc,
		OnDelete:     onDelete,
		OnUpdate:     onUpdate,
	}
	nw.ExtendBaseWidget(nw)
	return nw
}

// CreateRenderer implements the widget.Renderer interface.
func (nw *NodeWidget) CreateRenderer() fyne.WidgetRenderer {
	// Create circle representing the node.
	circle := canvas.NewCircle(color.RGBA{R: 0, G: 0, B: 255, A: 255})
	circle.StrokeWidth = 2
	circle.StrokeColor = color.White
	circle.FillColor = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	circle.Resize(fyne.NewSize(40, 40))
	circle.Move(fyne.NewPos(0, 0))

	// Create label to display node title.
	label := widget.NewLabel(nw.Node.Title)
	label.Move(fyne.NewPos(45, 10))
	label.Resize(fyne.NewSize(120, 20))

	objects := []fyne.CanvasObject{circle, label}
	return &nodeWidgetRenderer{objects: objects}
}

type nodeWidgetRenderer struct {
	objects []fyne.CanvasObject
}

func (r *nodeWidgetRenderer) Layout(_ fyne.Size) {}
func (r *nodeWidgetRenderer) MinSize() fyne.Size { return fyne.NewSize(160, 40) }
func (r *nodeWidgetRenderer) Refresh() {
	for _, obj := range r.objects {
		obj.Refresh()
	}
}
func (r *nodeWidgetRenderer) BackgroundColor() color.Color { return color.Transparent }
func (r *nodeWidgetRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *nodeWidgetRenderer) Destroy()                     {}

// Tapped opens an edit dialog for the node.
func (nw *NodeWidget) Tapped(_ *fyne.PointEvent) {
	var pop dialog.Dialog

	// Create form entries pre-populated with node data.
	titleEntry := widget.NewEntry()
	titleEntry.SetText(nw.Node.Title)
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetText(nw.Node.Content)
	typeSelect := widget.NewSelect([]string{
		string(domain.Concept), string(domain.Note), string(domain.References),
	}, nil)
	typeSelect.SetSelected(string(nw.Node.Type))
	tagsEntry := widget.NewEntry()
	tagsEntry.SetText(strings.Join(nw.Node.Tags, ", "))

	formItems := []*widget.FormItem{
		widget.NewFormItem("Title", titleEntry),
		widget.NewFormItem("Content", contentEntry),
		widget.NewFormItem("Type", typeSelect),
		widget.NewFormItem("Tags (comma separated)", tagsEntry),
	}
	form := widget.NewForm(formItems...)
	// Disable default buttons.
	form.SubmitText = ""
	form.CancelText = ""

	// Create custom buttons.
	updateBtn := widget.NewButton("Update", func() {
		nw.Node.Title = titleEntry.Text
		nw.Node.Content = contentEntry.Text
		nw.Node.Type = domain.NodeType(typeSelect.Selected)
		nw.Node.Tags = parseTags(tagsEntry.Text)
		err := nw.UseCase.UpdateNode(context.Background(), nw.Node)
		if err != nil {
			dialog.ShowError(err, nw.ParentWindow)
		} else {
			if nw.OnUpdate != nil {
				nw.OnUpdate(nw.Node)
			}
		}
		pop.Hide()
	})
	deleteBtn := widget.NewButton("Delete", func() {
		dialog.ShowConfirm("Delete Node", "Are you sure you want to delete this node?", func(confirm bool) {
			if confirm {
				err := nw.UseCase.DeleteNode(context.Background(), nw.Node.ID)
				if err != nil {
					dialog.ShowError(err, nw.ParentWindow)
				} else {
					if nw.OnDelete != nil {
						nw.OnDelete(nw.Node)
					}
				}
				pop.Hide()
			}
		}, nw.ParentWindow)
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		pop.Hide()
	})

	btnBar := container.New(layout.NewGridLayoutWithColumns(3), updateBtn, deleteBtn, cancelBtn)
	dialogContent := container.NewVBox(form, btnBar)

	pop = dialog.NewCustomWithoutButtons("Edit Node", dialogContent, nw.ParentWindow)
	pop.Show()
}

func (nw *NodeWidget) TappedSecondary(_ *fyne.PointEvent) {}

// buildGraphContainer builds and returns a new container with nodes and edges.
func buildGraphContainer(useCase *usecase.NodeUseCase, nodes []*domain.Node, edges []Edge, w fyne.Window, onDelete, onUpdate func(*domain.Node)) *fyne.Container {
	graph := container.NewWithoutLayout()
	positions := generatePositions(nodes, 100, 500, 500, 80)
	// Draw edges.
	for _, edge := range edges {
		if posFrom, ok1 := positions[edge.From.ID]; ok1 {
			if posTo, ok2 := positions[edge.To.ID]; ok2 {
				line := canvas.NewLine(color.Black)
				centerOffset := fyne.NewPos(20, 20)
				line.Position1 = posFrom.Add(centerOffset)
				line.Position2 = posTo.Add(centerOffset)
				line.StrokeWidth = 2
				graph.Add(line)
			}
		}
	}
	// Draw nodes.
	for _, n := range nodes {
		if pos, ok := positions[n.ID]; ok {
			nodeW := NewNodeWidget(n, pos, w, useCase, onDelete, onUpdate)
			nodeW.Move(pos)
			nodeW.Resize(fyne.NewSize(160, 40))
			graph.Add(nodeW)
		}
	}
	return graph
}

// ShowGraphUI displays the graph UI with search and management functionalities.
func ShowGraphUI(useCase *usecase.NodeUseCase, nodes []*domain.Node, initialEdges []Edge) {
	a := app.New()
	w := a.NewWindow("Neo4j Go Playground")
	w.Resize(fyne.NewSize(800, 600))

	// allNodes holds the complete list of nodes.
	allNodes := nodes
	// filteredNodes holds the current filtered list.
	filteredNodes := nodes
	// filteredEdges holds the current edges.
	filteredEdges := initialEdges
	// allEdges holds all relationships
	allEdges := initialEdges

	scrollContainer := container.NewScroll(container.NewWithoutLayout())
	scrollContainer.SetMinSize(fyne.NewSize(800, 600))

	var onDeleteCallback func(*domain.Node)
	var onUpdateCallback func(*domain.Node)

	// onDeleteCallback removes the deleted node and its edges and rebuilds the graph.
	onDeleteCallback = func(deletedNode *domain.Node) {
		allNodes = filterNodes(allNodes, func(n *domain.Node) bool { return n.ID != deletedNode.ID })
		filteredNodes = filterNodes(filteredNodes, func(n *domain.Node) bool { return n.ID != deletedNode.ID })
		filteredEdges = filterEdges(filteredEdges, deletedNode.ID)
		newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = newGraph
		scrollContainer.Refresh()
		w.Content().Refresh()
	}

	// onUpdateCallback rebuilds the graph after a node update.
	onUpdateCallback = func(updatedNode *domain.Node) {
		newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = newGraph
		scrollContainer.Refresh()
		w.Content().Refresh()
	}

	// Build initial graph.
	graphContainer := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
	scrollContainer.Content = graphContainer

	// --- Search UI ---
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Enter search query...")
	searchSelect := widget.NewSelect([]string{"Tag", "Title/Content", "All"}, nil)
	searchSelect.SetSelected("All")
	searchButton := widget.NewButton("Search", func() {
		if searchEntry.Text == "" {
			return
		}
		query := strings.TrimSpace(searchEntry.Text)
		criteria := searchSelect.Selected
		// Call the search method in the usecase layer.
		results, err := useCase.SearchNodes(context.Background(), query, criteria)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		// Update filteredNodes and rebuild graph accordingly.
		filteredNodes = results
		// For edges, you might want to decide how to update them (e.g., only show edges where both nodes are in results).
		var newEdges []Edge
		for _, e := range initialEdges {
			if containsNode(filteredNodes, e.From) && containsNode(filteredNodes, e.To) {
				newEdges = append(newEdges, e)
			}
		}
		filteredEdges = newEdges

		newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = newGraph
		scrollContainer.Refresh()
		w.Content().Refresh()
	})
	resetButton := widget.NewButton("Reset", func() {
		if searchEntry.Text == "" {
			return
		}
		filteredNodes = allNodes
		filteredEdges = allEdges
		newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = newGraph
		scrollContainer.Refresh()
		w.Content().Refresh()
		searchEntry.SetText("")
	})
	firstRow := container.NewBorder(nil, nil, searchSelect, nil, searchEntry)
	secondRow := container.NewAdaptiveGrid(2, searchButton, resetButton)
	searchContainer := container.NewVBox(firstRow, secondRow)

	// --- Top Buttons ---
	addNodeButton := widget.NewButton("Add Node", func() {
		titleEntry := widget.NewEntry()
		contentEntry := widget.NewMultiLineEntry()
		typeSelect := widget.NewSelect([]string{"CONCEPT", "NOTE", "REFERENCE"}, nil)
		typeSelect.SetSelected("CONCEPT")
		tagsEntry := widget.NewEntry()
		formItems := []*widget.FormItem{
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Content", contentEntry),
			widget.NewFormItem("Type", typeSelect),
			widget.NewFormItem("Tags (comma separated)", tagsEntry),
		}
		dialog.ShowForm("Add New Node", "Submit", "Cancel", formItems, func(valid bool) {
			if !valid {
				return
			}

			newNode := &domain.Node{
				Title:   titleEntry.Text,
				Content: contentEntry.Text,
				Type:    domain.NodeType(typeSelect.Selected),
				Tags:    parseTags(tagsEntry.Text),
			}

			go func() {
				id, err := useCase.CreateNode(context.Background(), newNode)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				newNode.ID = id
				allNodes = append(allNodes, newNode)
				filteredNodes = allNodes
				newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
				scrollContainer.Content = newGraph
				scrollContainer.Refresh()
				w.Content().Refresh()
			}()
		}, w)
	})
	addRelButton := widget.NewButton("Add Relationship", func() {
		sourceOptions := make([]string, len(allNodes))
		for i, n := range allNodes {
			sourceOptions[i] = n.Title
		}

		sourceSelect := widget.NewSelect(sourceOptions, nil)
		if len(sourceOptions) > 0 {
			sourceSelect.SetSelected(sourceOptions[0])
		}

		targetOptions := make([]string, len(allNodes))
		for i, n := range allNodes {
			targetOptions[i] = n.Title
		}

		targetCheckGroup := widget.NewCheckGroup(targetOptions, nil)
		relTypeSelect := widget.NewSelect([]string{
			string(domain.RelatedTo), string(domain.References), string(domain.IsPartOf),
			string(domain.HasPart), string(domain.DependsOn), string(domain.IsPrecededBy),
		}, nil)
		relTypeSelect.SetSelected(string(domain.RelatedTo))
		descEntry := widget.NewEntry()
		formItems := []*widget.FormItem{
			widget.NewFormItem("Source Node", sourceSelect),
			widget.NewFormItem("Target Nodes", targetCheckGroup),
			widget.NewFormItem("Relationship Type", relTypeSelect),
			widget.NewFormItem("Description", descEntry),
		}
		dialog.ShowForm("Add New Relationship", "Submit", "Cancel", formItems, func(valid bool) {
			if !valid {
				return
			}

			sourceIndex := indexOf(sourceOptions, sourceSelect.Selected)
			if sourceIndex < 0 {
				return
			}

			sourceNode := allNodes[sourceIndex]
			var targetIDs []string

			for _, sel := range targetCheckGroup.Selected {
				tIdx := indexOf(targetOptions, sel)
				if tIdx >= 0 && tIdx != sourceIndex {
					targetIDs = append(targetIDs, allNodes[tIdx].ID)
				}
			}

			if len(targetIDs) == 0 {
				return
			}

			newRel := &domain.Relationship{
				SourceID:    sourceNode.ID,
				TargetIDs:   targetIDs,
				Type:        domain.RelationType(relTypeSelect.Selected),
				Description: descEntry.Text,
			}

			go func() {
				createdIDs, err := useCase.CreateRelationship(context.Background(), newRel)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				if len(createdIDs) != len(targetIDs) {
					fmt.Printf("Warning: createdIDs length (%d) does not match targetIDs length (%d)\n", len(createdIDs), len(targetIDs))
				}

				for i, relID := range createdIDs {
					if i >= len(targetIDs) {
						break
					}

					tID := targetIDs[i]
					var targetNode *domain.Node
					for _, n := range allNodes {
						if n.ID == tID {
							targetNode = n
							break
						}
					}

					if targetNode != nil {
						newEdge := Edge{
							ID:   relID,
							From: sourceNode,
							To:   targetNode,
							Type: relTypeSelect.Selected,
						}

						allEdges = append(allEdges, newEdge)
						filteredEdges = append(filteredEdges, newEdge)
					}
				}

				newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
				scrollContainer.Content = newGraph
				scrollContainer.Refresh()
				w.Content().Refresh()
			}()
		}, w)
	})
	removeRelButton := widget.NewButton("Remove Relationship", func() {
		if len(filteredEdges) == 0 {
			dialog.ShowInformation("No relationships", "No relationships to remove", w)
			return
		}

		var edgeOptions []string
		for i, e := range filteredEdges {
			label := fmt.Sprintf("[%d] %s -> %s (%s)", i, e.From.Title, e.To.Title, e.Type)
			edgeOptions = append(edgeOptions, label)
		}

		edgeSelect := widget.NewSelect(edgeOptions, nil)
		if len(edgeOptions) > 0 {
			edgeSelect.SetSelected(edgeOptions[0])
		}

		formItems := []*widget.FormItem{
			widget.NewFormItem("Select Relationship", edgeSelect),
		}

		dialog.ShowForm("Remove Relationship", "Delete", "Cancel", formItems, func(valid bool) {
			if !valid {
				return
			}

			idx := indexOf(edgeOptions, edgeSelect.Selected)
			if idx < 0 {
				return
			}

			edge := filteredEdges[idx]
			if edge.ID != "" {
				err := useCase.DeleteRelationship(context.Background(), edge.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
			}

			filteredEdges = append(filteredEdges[:idx], filteredEdges[idx+1:]...)
			newGraph := buildGraphContainer(useCase, filteredNodes, filteredEdges, w, onDeleteCallback, onUpdateCallback)
			scrollContainer.Content = newGraph
			scrollContainer.Refresh()
			w.Content().Refresh()
		}, w)
	})

	topButtons := container.NewAdaptiveGrid(3, addNodeButton, addRelButton, removeRelButton)
	content := container.NewBorder(searchContainer, topButtons, nil, nil, scrollContainer)
	w.SetContent(content)
	w.ShowAndRun()
}

// Helper function: generate random positions for nodes.
func generatePositions(nodes []*domain.Node, minDist float64, width, height, maxAttempts int) map[string]fyne.Position {
	randomize := rand.New(rand.NewSource(time.Now().UnixNano()))
	positions := make(map[string]fyne.Position, len(nodes))

	for _, n := range nodes {
		var pos fyne.Position
		attempt := 0

		for {
			attempt++
			x := randomize.Intn(width) + 50
			y := randomize.Intn(height) + 50
			candidate := fyne.NewPos(float32(x), float32(y))
			if !tooClose(candidate, positions, minDist) {
				pos = candidate
				break
			}

			if attempt > maxAttempts {
				pos = candidate
				break
			}
		}
		positions[n.ID] = pos
	}
	return positions
}

// Helper function: check if candidate position is too close to existing ones.
func tooClose(candidate fyne.Position, existing map[string]fyne.Position, minDist float64) bool {
	for _, pos := range existing {
		if distance(candidate, pos) < minDist {
			return true
		}
	}
	return false
}

// Helper function: calculate Euclidean distance.
func distance(a, b fyne.Position) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// Helper function: parse comma-separated tags.
func parseTags(tagsStr string) []string {
	tags := strings.Split(tagsStr, ",")
	var result []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper function: return index of a string in a slice.
func indexOf(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

// Helper function: filter nodes based on predicate.
func filterNodes(nodes []*domain.Node, predicate func(*domain.Node) bool) []*domain.Node {
	var result []*domain.Node
	for _, n := range nodes {
		if predicate(n) {
			result = append(result, n)
		}
	}
	return result
}

// Helper function: filter edges by excluding those referencing a given node ID.
func filterEdges(edges []Edge, nodeID string) []Edge {
	var result []Edge
	for _, e := range edges {
		if e.From.ID != nodeID && e.To.ID != nodeID {
			result = append(result, e)
		}
	}
	return result
}

// Helper function: check if a node is in the list.
func containsNode(list []*domain.Node, node *domain.Node) bool {
	for _, n := range list {
		if n.ID == node.ID {
			return true
		}
	}
	return false
}
