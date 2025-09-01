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
	app        *tview.Application
	flex       *tview.Flex
	scrollView *tview.Flex
	panels     []*tview.TextView
	timeRange  *tview.TextView
	focusIndex int
	histories  []*QueryHistory
	onQuit     func()
}

// NewTUI creates a new terminal user interface
func NewTUI(queries []backend.Query, onQuit func()) *TUI {
	tui := &TUI{
		app:       tview.NewApplication(),
		histories: make([]*QueryHistory, len(queries)),
		onQuit:    onQuit,
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

	for i, query := range queries {
		panel := tview.NewTextView()
		panel.SetTitle(fmt.Sprintf(" %s ", query.Name))
		panel.SetBorder(true)
		panel.SetText("Initializing...")
		panel.SetDynamicColors(true)
		panel.SetWordWrap(false)

		t.panels[i] = panel
		// Each panel takes equal width, with minimum of 1/3 screen for 3+ panels
		if len(queries) <= 2 {
			t.scrollView.AddItem(panel, 0, 1, i == 0)
		} else {
			// For 3+ panels, make them wider so they need horizontal scrolling
			t.scrollView.AddItem(panel, 80, 0, i == 0)
		}
	}

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

// focusNext moves focus to the next panel
func (t *TUI) focusNext() {
	if len(t.panels) > 0 {
		t.focusIndex = (t.focusIndex + 1) % len(t.panels)
		t.updateFocus()
	}
}

// focusPrev moves focus to the previous panel
func (t *TUI) focusPrev() {
	if len(t.panels) > 0 {
		t.focusIndex = (t.focusIndex - 1 + len(t.panels)) % len(t.panels)
		t.updateFocus()
	}
}

// updateFocus updates the visual focus indicator
func (t *TUI) updateFocus() {
	if len(t.panels) == 0 {
		return
	}

	for i, panel := range t.panels {
		if i == t.focusIndex {
			panel.SetBorderColor(tcell.ColorYellow)
			panel.SetTitleColor(tcell.ColorYellow)
		} else {
			panel.SetBorderColor(tcell.ColorWhite)
			panel.SetTitleColor(tcell.ColorWhite)
		}
	}
	// Ensure focused panel is visible by setting focus
	t.app.SetFocus(t.panels[t.focusIndex])
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
	graphWidth := width - 7   // Leave margin
	graphHeight := height - 6 // Leave space for title and current value

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
