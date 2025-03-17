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

	"github.com/AndrivA89/knowledge-manager/internal/domain"
	"github.com/AndrivA89/knowledge-manager/internal/usecase"
)

type Edge struct {
	ID   string
	From *domain.Node
	To   *domain.Node
	Type string
}

type NodeWidget struct {
	widget.BaseWidget
	Node         *domain.Node
	Pos          fyne.Position
	ParentWindow fyne.Window
	UseCase      *usecase.NodeUseCase
	OnDelete     func(*domain.Node)
	OnUpdate     func(*domain.Node)
}

func NewNodeWidget(
	n *domain.Node,
	pos fyne.Position,
	w fyne.Window,
	uc *usecase.NodeUseCase,
	onDelete func(*domain.Node),
	onUpdate func(*domain.Node),
) *NodeWidget {
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

func (nw *NodeWidget) CreateRenderer() fyne.WidgetRenderer {
	circle := canvas.NewCircle(color.RGBA{R: 0, G: 0, B: 255, A: 255})
	circle.StrokeWidth = 2
	circle.StrokeColor = color.White
	circle.FillColor = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	circle.Resize(fyne.NewSize(40, 40))
	circle.Move(fyne.NewPos(0, 0))

	label := widget.NewLabel(nw.Node.Title)
	label.Move(fyne.NewPos(45, 10))
	label.Resize(fyne.NewSize(120, 20))

	objects := []fyne.CanvasObject{circle, label}
	return &nodeWidgetRenderer{objects: objects}
}

type nodeWidgetRenderer struct {
	objects []fyne.CanvasObject
}

func (r *nodeWidgetRenderer) Layout(size fyne.Size) {}

func (r *nodeWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(160, 40)
}

func (r *nodeWidgetRenderer) Refresh() {
	for _, obj := range r.objects {
		obj.Refresh()
	}
}

func (r *nodeWidgetRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *nodeWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *nodeWidgetRenderer) Destroy() {}

func (nw *NodeWidget) Tapped(_ *fyne.PointEvent) {
	var pop dialog.Dialog

	titleEntry := widget.NewEntry()
	titleEntry.SetText(nw.Node.Title)
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetText(nw.Node.Content)
	typeSelect := widget.NewSelect([]string{
		string(domain.Concept), string(domain.Note), string(domain.References)},
		nil,
	)
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
	form.SubmitText = ""
	form.CancelText = ""

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
			pop.Hide()
		}
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

func buildGraphContainer(
	useCase *usecase.NodeUseCase,
	nodes []*domain.Node,
	edges []Edge,
	w fyne.Window,
	onDelete func(*domain.Node),
	onUpdate func(*domain.Node),
) *fyne.Container {
	containerGraph := container.NewWithoutLayout()
	positions := generatePositions(nodes, 100, 500, 500, 80)

	for _, edge := range edges {
		posFrom, ok1 := positions[edge.From.ID]
		posTo, ok2 := positions[edge.To.ID]
		if ok1 && ok2 {
			line := canvas.NewLine(color.Black)
			centerOffset := fyne.NewPos(20, 20)
			line.Position1 = posFrom.Add(centerOffset)
			line.Position2 = posTo.Add(centerOffset)
			line.StrokeWidth = 2
			containerGraph.Add(line)
		}
	}

	for _, n := range nodes {
		pos, ok := positions[n.ID]
		if !ok {
			continue
		}
		nodeW := NewNodeWidget(n, pos, w, useCase, onDelete, onUpdate)
		nodeW.Move(pos)
		nodeW.Resize(fyne.NewSize(160, 40))
		containerGraph.Add(nodeW)
	}
	return containerGraph
}

func ShowGraphUI(useCase *usecase.NodeUseCase, nodes []*domain.Node, initialEdges []Edge) {
	a := app.New()
	w := a.NewWindow("Graph Visualization")
	w.Resize(fyne.NewSize(800, 600))

	edges := initialEdges
	scrollContainer := container.NewScroll(container.NewWithoutLayout())
	scrollContainer.SetMinSize(fyne.NewSize(800, 600))

	var onDeleteCallback func(*domain.Node)
	var onUpdateCallback func(*domain.Node)

	onUpdateCallback = func(updatedNode *domain.Node) {
		graphContainer := buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = graphContainer
		scrollContainer.Refresh()
		w.Content().Refresh()
	}

	onDeleteCallback = func(deletedNode *domain.Node) {
		for i, node := range nodes {
			if node.ID == deletedNode.ID {
				nodes = append(nodes[:i], nodes[i+1:]...)
				break
			}
		}
		newEdges := make([]Edge, 0, len(edges))
		for _, e := range edges {
			if e.From.ID != deletedNode.ID && e.To.ID != deletedNode.ID {
				newEdges = append(newEdges, e)
			}
		}
		edges = newEdges

		graphContainer := buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
		scrollContainer.Content = graphContainer
		scrollContainer.Refresh()
		w.Content().Refresh()
	}

	graphContainer := buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
	scrollContainer.Content = graphContainer
	scrollContainer.Refresh()

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
				nodes = append(nodes, newNode)
				graphContainer = buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
				scrollContainer.Content = graphContainer
				scrollContainer.Refresh()
				w.Content().Refresh()
			}()
		}, w)
	})

	addRelButton := widget.NewButton("Add Relationship", func() {
		sourceOptions := make([]string, len(nodes))
		for i, n := range nodes {
			sourceOptions[i] = n.Title
		}
		sourceSelect := widget.NewSelect(sourceOptions, nil)
		if len(sourceOptions) > 0 {
			sourceSelect.SetSelected(sourceOptions[0])
		}

		targetOptions := make([]string, len(nodes))
		for i, n := range nodes {
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
			sourceNode := nodes[sourceIndex]

			var targetIDs []string
			for _, sel := range targetCheckGroup.Selected {
				tIdx := indexOf(targetOptions, sel)
				if tIdx >= 0 && tIdx != sourceIndex {
					targetIDs = append(targetIDs, nodes[tIdx].ID)
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
					for _, n := range nodes {
						if n.ID == tID {
							targetNode = n
							break
						}
					}
					if targetNode != nil {
						edges = append(edges, Edge{
							ID:   relID,
							From: sourceNode,
							To:   targetNode,
							Type: relTypeSelect.Selected,
						})
					}
				}
				graphContainer = buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
				scrollContainer.Content = graphContainer
				scrollContainer.Refresh()
				w.Content().Refresh()
			}()
		}, w)
	})

	removeRelButton := widget.NewButton("Remove Relationship", func() {
		if len(edges) == 0 {
			dialog.ShowInformation("No relationships", "No relationships to remove", w)
			return
		}
		var edgeOptions []string
		for i, e := range edges {
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
			edge := edges[idx]
			if edge.ID != "" {
				err := useCase.DeleteRelationship(context.Background(), edge.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
			}
			edges = append(edges[:idx], edges[idx+1:]...)
			graphContainer = buildGraphContainer(useCase, nodes, edges, w, onDeleteCallback, onUpdateCallback)
			scrollContainer.Content = graphContainer
			scrollContainer.Refresh()
			w.Content().Refresh()
		}, w)
	})

	buttons := container.NewHBox(addNodeButton, addRelButton, removeRelButton)
	content := container.NewBorder(buttons, nil, nil, nil, scrollContainer)
	w.SetContent(content)
	w.ShowAndRun()
}

func generatePositions(nodes []*domain.Node, minDist float64, width, height, maxAttempts int) map[string]fyne.Position {
	rand.Seed(time.Now().UnixNano())
	positions := make(map[string]fyne.Position, len(nodes))
	for _, n := range nodes {
		var pos fyne.Position
		attempt := 0
		for {
			attempt++
			x := rand.Intn(width) + 50
			y := rand.Intn(height) + 50
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

func tooClose(candidate fyne.Position, existing map[string]fyne.Position, minDist float64) bool {
	for _, pos := range existing {
		if distance(candidate, pos) < minDist {
			return true
		}
	}
	return false
}

func distance(a, b fyne.Position) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

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

func indexOf(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}
