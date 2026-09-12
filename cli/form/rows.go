package form

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// Row is one entry of a Rows input, the tag holds the original value of
// rows loaded from the resource and is nil for added rows.
type Row struct {
	Cells []Input
	Tag   interface{}
}

// rowHit is the rendered position of a cell or button for mouse clicks.
type rowHit struct {
	y   int
	x   int
	w   int
	h   int
	row int
	col int
}

func (h rowHit) contains(x, y int) bool {
	return y >= h.y && y < h.y+h.h && x >= h.x && x < h.x+h.w
}

// Rows is a list of multi input rows such as node ports, each row shows
// its cells one per line with a label followed by the remove and add
// buttons, rows are separated by a divider. The focus moves through the
// cells and buttons of every row.
type Rows struct {
	Base

	// Empty is shown with the add button when there are no rows.
	Empty    string
	AddLabel string

	// Labels are the labels of the cells of a row.
	Labels []string

	// New creates an added row.
	New func() *Row

	rows    []*Row
	removed []*Row
	initial int

	focused bool
	row     int
	col     int
	hits    []rowHit
}

func NewRows(label, empty, addLabel string, labels []string,
	newRow func() *Row) *Rows {

	return &Rows{
		Base:     Base{Label: label},
		Empty:    empty,
		AddLabel: addLabel,
		Labels:   labels,
		New:      newRow,
		rows:     []*Row{},
		removed:  []*Row{},
		row:      -1,
	}
}

// Add appends an initial row loaded from the resource.
func (r *Rows) Add(row *Row) {
	r.rows = append(r.rows, row)
	r.initial = len(r.rows)
}

func (r *Rows) Rows() []*Row {
	return r.rows
}

// Removed returns the tags of the removed initial rows.
func (r *Rows) Removed() []interface{} {
	tags := []interface{}{}
	for _, row := range r.removed {
		tags = append(tags, row.Tag)
	}
	return tags
}

func (r *Rows) Changed() bool {
	if len(r.rows) != r.initial || len(r.removed) > 0 {
		return true
	}
	for _, row := range r.rows {
		for _, cell := range row.Cells {
			if cell.Changed() {
				return true
			}
		}
	}
	return false
}

// cellLabel returns the label of the cell at the index.
func (r *Rows) cellLabel(col int) string {
	if col < len(r.Labels) {
		return r.Labels[col]
	}
	return ""
}

// labelWidth is the width of the label column of the cells.
func (r *Rows) labelWidth() int {
	width := 0
	for _, label := range r.Labels {
		width = max(width, len([]rune(label)))
	}
	if width == 0 {
		return 0
	}
	return width + 2
}

// cell returns the focused cell input or nil when a button is focused.
func (r *Rows) cell() Input {
	if r.row < 0 || r.row >= len(r.rows) {
		return nil
	}
	if r.col < 0 || r.col >= len(r.rows[r.row].Cells) {
		return nil
	}
	return r.rows[r.row].Cells[r.col]
}

func (r *Rows) removeCol(row int) int {
	return len(r.rows[row].Cells)
}

func (r *Rows) addCol(row int) int {
	return len(r.rows[row].Cells) + 1
}

func (r *Rows) focusCell() tea.Cmd {
	cell := r.cell()
	if cell == nil {
		return nil
	}
	return cell.Focus(false)
}

func (r *Rows) blurCell() {
	cell := r.cell()
	if cell != nil {
		cell.Blur()
	}
}

func (r *Rows) Focused() bool {
	return r.focused
}

func (r *Rows) Focus(reverse bool) tea.Cmd {
	r.focused = true

	if len(r.rows) == 0 {
		r.row = -1
		r.col = 0
		return nil
	}

	if reverse {
		r.row = len(r.rows) - 1
		r.col = r.addCol(r.row)
	} else {
		r.row = 0
		r.col = 0
	}

	return r.focusCell()
}

func (r *Rows) Blur() {
	r.blurCell()
	r.focused = false
}

// Next moves through the cells and buttons of the rows.
func (r *Rows) Next(reverse bool) (tea.Cmd, bool) {
	if len(r.rows) == 0 || r.row < 0 {
		return nil, false
	}

	r.blurCell()

	if reverse {
		r.col--
		if r.col < 0 {
			r.row--
			if r.row < 0 {
				r.row = 0
				r.col = 0
				return nil, false
			}
			r.col = r.addCol(r.row)
		}
	} else {
		r.col++
		if r.col > r.addCol(r.row) {
			r.row++
			if r.row >= len(r.rows) {
				r.row = len(r.rows) - 1
				r.col = r.addCol(r.row)
				return nil, false
			}
			r.col = 0
		}
	}

	return r.focusCell(), true
}

// remove removes the row keeping the tag of initial rows.
func (r *Rows) remove(index int) tea.Cmd {
	if index < 0 || index >= len(r.rows) {
		return nil
	}

	r.blurCell()

	row := r.rows[index]
	if row.Tag != nil {
		r.removed = append(r.removed, row)
	}
	r.rows = append(r.rows[:index], r.rows[index+1:]...)

	if len(r.rows) == 0 {
		r.row = -1
		r.col = 0
		return nil
	}

	r.row = min(index, len(r.rows)-1)
	r.col = 0
	return r.focusCell()
}

// add inserts a new row after the index, -1 appends.
func (r *Rows) add(index int) tea.Cmd {
	if r.New == nil {
		return nil
	}

	r.blurCell()

	row := r.New()
	index++
	if index < 0 || index > len(r.rows) {
		index = len(r.rows)
	}

	r.rows = append(r.rows, nil)
	copy(r.rows[index+1:], r.rows[index:])
	r.rows[index] = row

	r.row = index
	r.col = 0
	return r.focusCell()
}

func (r *Rows) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !r.focused {
		return nil, false
	}

	cell := r.cell()
	if cell != nil {
		cmd, consumed := cell.Update(msg)
		if consumed {
			return cmd, true
		}
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok || !keyIs(keyMsg, "enter", "space") {
		return nil, false
	}

	if r.row < 0 {
		return r.add(-1), true
	}
	if r.col == r.removeCol(r.row) {
		return r.remove(r.row), true
	}
	if r.col == r.addCol(r.row) {
		return r.add(r.row), true
	}

	return nil, false
}

// Click focuses the cell or presses the button at the position.
func (r *Rows) Click(x, y int) tea.Cmd {
	for _, hit := range r.hits {
		if !hit.contains(x, y) {
			continue
		}

		r.blurCell()
		r.focused = true
		r.row = hit.row
		r.col = hit.col

		if hit.row < 0 {
			return r.add(-1)
		}
		if hit.col == r.removeCol(hit.row) {
			return r.remove(hit.row)
		}
		if hit.col == r.addCol(hit.row) {
			return r.add(hit.row)
		}

		cell := r.cell()
		if cell == nil {
			return nil
		}
		return tea.Batch(
			cell.Focus(false),
			cell.Click(x-hit.x-r.labelWidth(), y-hit.y),
		)
	}

	return nil
}

func (r *Rows) renderButton(label string, style lipgloss.Style,
	focused bool) string {

	if focused {
		return buttonFocusStyle.Render(label)
	}
	return style.Render(label)
}

func (r *Rows) View(width int) string {
	r.hits = []rowHit{}
	lines := []string{renderLabel(r.Label, r.focused)}
	y := 1

	if len(r.rows) == 0 {
		btn := r.renderButton("+ "+r.AddLabel, addButtonStyle,
			r.focused && r.row < 0)
		r.hits = append(r.hits, rowHit{
			y:   y,
			x:   0,
			w:   lipgloss.Width(btn),
			h:   1,
			row: -1,
			col: 0,
		})
		lines = append(lines, btn)
		return strings.Join(lines, "\n")
	}

	labelWidth := r.labelWidth()
	cellWidth := max(width-labelWidth, 10)
	divider := dividerStyle.Render(strings.Repeat("─", max(width, 1)))

	for ri, row := range r.rows {
		if ri > 0 {
			lines = append(lines, divider)
			y++
		}

		focusRow := r.focused && r.row == ri

		// Cells stack one per line with their label, cells such as
		// token lists may span several lines
		for col, cell := range row.Cells {
			label := renderLabel(
				widget.Truncate(max(labelWidth-1, 0), "%s", r.cellLabel(col)),
				focusRow && r.col == col,
			)
			// Text fields get the panel background, selects and token
			// lists draw their own
			style := lipgloss.NewStyle()
			switch cell.(type) {
			case *Text, *Number:
				style = cellStyle
			}
			view := style.Width(cellWidth).MaxWidth(cellWidth).Render(
				cell.View(cellWidth))
			line := lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().Width(labelWidth).Render(label),
				view,
			)

			r.hits = append(r.hits, rowHit{
				y:   y,
				x:   0,
				w:   width,
				h:   lipgloss.Height(line),
				row: ri,
				col: col,
			})
			lines = append(lines, line)
			y += lipgloss.Height(line)
		}

		remove := r.renderButton("- Remove", removeButtonStyle,
			focusRow && r.col == r.removeCol(ri))
		add := r.renderButton("+ Add", addButtonStyle,
			focusRow && r.col == r.addCol(ri))

		r.hits = append(r.hits, rowHit{
			y:   y,
			x:   labelWidth,
			w:   lipgloss.Width(remove),
			h:   1,
			row: ri,
			col: r.removeCol(ri),
		}, rowHit{
			y:   y,
			x:   labelWidth + lipgloss.Width(remove) + 1,
			w:   lipgloss.Width(add),
			h:   1,
			row: ri,
			col: r.addCol(ri),
		})
		lines = append(lines, strings.Repeat(" ", labelWidth)+
			remove+" "+add)
		y++
	}

	return strings.Join(lines, "\n")
}
