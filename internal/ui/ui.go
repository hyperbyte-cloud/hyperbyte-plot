package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/guptarohit/asciigraph"
	"github.com/rivo/tview"

	"promviz/internal/backend"
)

// QueryHistory maintains time series data for a single query
type QueryHistory struct {
	Name       string
	TimeSeries *backend.TimeSeriesResult
	LastError  error
}

// TUI represents the terminal user interface
type TUI struct {
	app           *tview.Application
	flex          *tview.Flex
	scrollView    *tview.Flex
	panels        []*tview.TextView
	timeRange     *tview.TextView
	focusIndex    int
	scrollOffset  int // Track horizontal scroll position
	visiblePanels int // Number of panels visible at once
	histories     []*QueryHistory
	onQuit        func()
}

// NewTUI creates a new terminal user interface
func NewTUI(queries []backend.Query, onQuit func()) *TUI {
	tui := &TUI{
		app:           tview.NewApplication(),
		histories:     make([]*QueryHistory, len(queries)),
		onQuit:        onQuit,
		focusIndex:    0,
		scrollOffset:  0,
		visiblePanels: 3, // Default to showing 3 panels at once
	}

	// Initialize query histories
	for i, query := range queries {
		tui.histories[i] = &QueryHistory{
			Name:       query.Name,
			TimeSeries: &backend.TimeSeriesResult{Points: []backend.DataPoint{}},
			LastError:  nil,
		}
	}

	tui.setupUI(queries)
	return tui
}

// setupUI initializes the TUI layout with horizontal scrolling
func (t *TUI) setupUI(queries []backend.Query) {
	// Create main vertical container
	t.flex = tview.NewFlex().SetDirection(tview.FlexRow)

	// Create horizontal scrollable container for panels
	t.scrollView = tview.NewFlex().SetDirection(tview.FlexColumn)
	t.panels = make([]*tview.TextView, len(queries))

	// Create all panels but don't add them to scrollView yet
	for i, query := range queries {
		panel := tview.NewTextView()
		panel.SetTitle(fmt.Sprintf(" %s ", query.Name))
		panel.SetBorder(true)
		panel.SetText("Initializing...")
		panel.SetDynamicColors(true)
		panel.SetWordWrap(false)

		t.panels[i] = panel
	}

	// Adjust visible panels based on total number
	if len(queries) <= 2 {
		t.visiblePanels = len(queries)
	} else if len(queries) == 3 {
		t.visiblePanels = 3
	} else {
		t.visiblePanels = 3 // Show max 3 panels at once for 4+ queries
	}

	// Initialize the scroll view with visible panels
	t.updateScrollView()

	// Add time range display at the bottom
	t.timeRange = tview.NewTextView()
	t.timeRange.SetText("Time Range: Waiting for data...")
	t.timeRange.SetTextAlign(tview.AlignCenter)
	t.timeRange.SetDynamicColors(true)

	// Add instructions at the very bottom
	instructions := tview.NewTextView()
	instructions.SetText("Navigation: ← → Arrow keys or Tab/Shift+Tab to switch panels | q/Q to quit")
	instructions.SetTextAlign(tview.AlignCenter)
	instructions.SetDynamicColors(true)

	// Add scrollable view, time range, and instructions to main container
	t.flex.AddItem(t.scrollView, 0, 1, true)
	t.flex.AddItem(t.timeRange, 1, 0, false)
	t.flex.AddItem(instructions, 1, 0, false)

	// Set up key bindings
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				if t.onQuit != nil {
					t.onQuit()
				}
				return nil
			}
		case tcell.KeyTab, tcell.KeyRight:
			t.focusNext()
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			t.focusPrev()
			return nil
		}
		return event
	})

	t.app.SetRoot(t.flex, true)
	t.updateFocus()
}

// updateScrollView refreshes the scroll view to show the correct panels
func (t *TUI) updateScrollView() {
	// Clear the current scroll view
	t.scrollView.Clear()

	if len(t.panels) == 0 {
		return
	}

	// Calculate which panels should be visible
	maxOffset := len(t.panels) - t.visiblePanels
	if maxOffset < 0 {
		maxOffset = 0
	}

	// Ensure scroll offset is within bounds
	if t.scrollOffset > maxOffset {
		t.scrollOffset = maxOffset
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}

	// Add visible panels to the scroll view
	endIndex := t.scrollOffset + t.visiblePanels
	if endIndex > len(t.panels) {
		endIndex = len(t.panels)
	}

	for i := t.scrollOffset; i < endIndex; i++ {
		panel := t.panels[i]
		t.scrollView.AddItem(panel, 0, 1, i == t.focusIndex)
	}
}

// scrollToShowFocus adjusts scroll offset to ensure focused panel is visible
func (t *TUI) scrollToShowFocus() {
	if len(t.panels) == 0 {
		return
	}

	// Check if focused panel is visible in current scroll view
	visibleStart := t.scrollOffset
	visibleEnd := t.scrollOffset + t.visiblePanels - 1

	// If focus is to the right of visible area, scroll right
	if t.focusIndex > visibleEnd {
		t.scrollOffset = t.focusIndex - t.visiblePanels + 1
	}
	// If focus is to the left of visible area, scroll left
	if t.focusIndex < visibleStart {
		t.scrollOffset = t.focusIndex
	}

	// Update the scroll view with new offset
	t.updateScrollView()
}

// focusNext moves focus to the next panel
func (t *TUI) focusNext() {
	if len(t.panels) > 0 {
		t.focusIndex = (t.focusIndex + 1) % len(t.panels)
		t.scrollToShowFocus()
		t.updateFocus()
	}
}

// focusPrev moves focus to the previous panel
func (t *TUI) focusPrev() {
	if len(t.panels) > 0 {
		t.focusIndex = (t.focusIndex - 1 + len(t.panels)) % len(t.panels)
		t.scrollToShowFocus()
		t.updateFocus()
	}
}

// updateFocus updates the visual focus indicator
func (t *TUI) updateFocus() {
	if len(t.panels) == 0 {
		return
	}

	// Update border colors for all panels
	for i, panel := range t.panels {
		if i == t.focusIndex {
			panel.SetBorderColor(tcell.ColorYellow)
			panel.SetTitleColor(tcell.ColorYellow)
		} else {
			panel.SetBorderColor(tcell.ColorWhite)
			panel.SetTitleColor(tcell.ColorWhite)
		}
	}

	// Set app focus to the focused panel (if it's currently visible)
	visibleStart := t.scrollOffset
	visibleEnd := t.scrollOffset + t.visiblePanels - 1
	if t.focusIndex >= visibleStart && t.focusIndex <= visibleEnd {
		t.app.SetFocus(t.panels[t.focusIndex])
	}
}

// updateTimeRange updates the time range display based on current data
func (t *TUI) updateTimeRange() {
	if t.timeRange == nil {
		return
	}

	var earliestTime, latestTime *time.Time
	hasData := false

	// Find the overall time range across all queries
	for _, history := range t.histories {
		if history.TimeSeries != nil && len(history.TimeSeries.Points) > 0 {
			for _, point := range history.TimeSeries.Points {
				if !hasData {
					earliestTime = &point.Timestamp
					latestTime = &point.Timestamp
					hasData = true
				} else {
					if point.Timestamp.Before(*earliestTime) {
						earliestTime = &point.Timestamp
					}
					if point.Timestamp.After(*latestTime) {
						latestTime = &point.Timestamp
					}
				}
			}
		}
	}

	var timeRangeText string
	if hasData && earliestTime != nil && latestTime != nil {
		timeRangeText = fmt.Sprintf("[yellow]Time Range:[white] %s [gray]to[white] %s",
			earliestTime.Format("15:04:05"),
			latestTime.Format("15:04:05"))
	} else {
		timeRangeText = "[gray]Time Range: Waiting for data...[white]"
	}

	t.timeRange.SetText(timeRangeText)
}

// UpdateTimeSeries updates a specific metric panel with new time series data
func (t *TUI) UpdateTimeSeries(index int, timeSeries *backend.TimeSeriesResult, err error) {
	if index < 0 || index >= len(t.histories) {
		return
	}

	// Update history data immediately (for tests)
	if err != nil {
		t.histories[index].LastError = err
	} else {
		t.histories[index].TimeSeries = timeSeries
		t.histories[index].LastError = nil
	}

	// Only queue UI updates if the app is properly initialized
	if t.app != nil && len(t.panels) > index {
		t.app.QueueUpdateDraw(func() {
			if err != nil {
				t.panels[index].SetText(fmt.Sprintf("[red]Error: %v[white]", err))
			} else {
				// Render the time series graph
				t.renderTimeSeriesGraph(index)
			}

			// Update the time range display
			t.updateTimeRange()
		})
	}
}

// renderTimeSeriesGraph renders a time series graph for the given panel
func (t *TUI) renderTimeSeriesGraph(index int) {
	history := t.histories[index]
	panel := t.panels[index]

	if len(history.TimeSeries.Points) == 0 {
		panel.SetText("No data available")
		return
	}

	// Sort points by timestamp to ensure correct order
	points := make([]backend.DataPoint, len(history.TimeSeries.Points))
	copy(points, history.TimeSeries.Points)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})

	// Extract values for graphing
	values := make([]float64, len(points))
	for i, point := range points {
		values[i] = point.Value
	}

	// Get panel dimensions dynamically
	_, _, width, height := panel.GetInnerRect()

	// Calculate graph dimensions (leave space for text)
	// Calculate margin based on max y value digits + 4 for outline space
	maxY := values[0]
	minY := values[0]
	for _, v := range values {
		if v > maxY {
			maxY = v
		}
		if v < minY {
			minY = v
		}
	}
	// Find the largest absolute value for y-axis
	absMaxY := maxY
	if -minY > maxY {
		absMaxY = -minY
	}
	yDigits := len(fmt.Sprintf("%.0f", absMaxY))
	if absMaxY < 0 {
		yDigits++ // for negative sign
	}
	margin := yDigits + 7
	graphWidth := width - margin // Leave margin based on y-axis label width
	graphHeight := height - 6    // Leave space for title and current value

	// Ensure minimum dimensions
	if graphWidth < 20 {
		graphWidth = 20
	}
	if graphHeight < 3 {
		graphHeight = 3
	}

	// Generate ASCII graph with dynamic sizing
	graph := asciigraph.Plot(values,
		asciigraph.Height(graphHeight),
		asciigraph.Width(graphWidth),
		asciigraph.Caption(fmt.Sprintf("%s Time Series", history.Name)))

	// Get latest value and timestamp
	latest := points[len(points)-1]

	// Create time range info
	oldest := points[0]
	timeRange := fmt.Sprintf("%s to %s",
		oldest.Timestamp.Format("15:04:05"),
		latest.Timestamp.Format("15:04:05"))

	// Build content with current value, time range, and graph
	content := fmt.Sprintf("[yellow]Current: %.2f[white]\n[gray]Time Range: %s[white]\n\n%s",
		latest.Value,
		timeRange,
		graph)

	panel.SetText(content)
}

// UpdateMetric maintains compatibility with old interface (deprecated)
func (t *TUI) UpdateMetric(index int, result backend.DataPoint, err error) {
	// Convert single result to time series for backward compatibility
	if err != nil {
		t.UpdateTimeSeries(index, nil, err)
		return
	}

	timeSeries := &backend.TimeSeriesResult{
		Points: []backend.DataPoint{result},
	}
	t.UpdateTimeSeries(index, timeSeries, nil)
}

// Run starts the TUI application
func (t *TUI) Run() error {
	return t.app.Run()
}

// Stop stops the TUI application
func (t *TUI) Stop() {
	t.app.Stop()
}
