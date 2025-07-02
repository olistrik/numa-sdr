// ntcharts - Copyright (c) 2024 Neomantra Corp.

// Package wavelinechart implements a linechart that draws wave lines on the graph
package spectrum

import (
	"math"
	"sort"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/buffer"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/olistrik/numa-sdr/api/unit"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func FrequencyLabelFormatter() linechart.LabelFormatter {
	return func(i int, v float64) string {
		return unit.Frequency(v).String()
	}
}

func MagnitudeLabelFormatter() linechart.LabelFormatter {
	return func(i int, v float64) string {
		return unit.Decabel(v).String()
	}
}

const DefaultDataSetName = "default"

type dataSet struct {
	LineStyle runes.LineStyle // type of line runes to draw
	Style     lipgloss.Style

	// stores data points from Plot() and contains scaled data points
	pBuf *buffer.Float64PointScaleBuffer
}

// Model contains state of a wavelinechart with an embedded linechart.Model
// A data set is a list of (X,Y) Cartesian coordinates.
// For each data set, wavelinecharts can only plot a single rune in each column
// of the graph canvas by mapping (X,Y) data points values in Cartesian coordinates
// to the (X,Y) canvas coordinates of the graph.
// If multiple data points map to the same column, then the latest data point
// will be used for that column.
// By default, there is a line through the graph X axis without any plotted points.
// Uses linechart Model UpdateHandler() for processing keyboard and mouse messages.
type Model struct {
	linechart.Model
	dLineStyle runes.LineStyle     // default data set LineStyletype
	dStyle     lipgloss.Style      // default data set Style
	dSets      map[string]*dataSet // maps names to data sets
}

// New returns a wavelinechart Model initialized
// with given linechart Model and various options.
// By default, the chart will auto set X and Y value ranges,
// and only enable moving viewport on X axis.
func New(w, h int, opts ...Option) Model {
	m := Model{
		Model: linechart.New(w, h, 0, 1, 0, 1,
			linechart.WithAutoXYRange(), // automatically adjust value ranges
			linechart.WithXYSteps(5, 5),
			linechart.WithXLabelFormatter(FrequencyLabelFormatter()),
			linechart.WithYLabelFormatter(MagnitudeLabelFormatter()),
			linechart.WithUpdateHandler(linechart.XAxisUpdateHandler(float64(1*unit.KHz))), // only scroll on X axis
		),
		dLineStyle: runes.ArcLineStyle,
		dStyle:     lipgloss.NewStyle(),
		dSets:      make(map[string]*dataSet),
	}
	for _, opt := range opts {
		opt(&m)
	}
	m.UpdateGraphSizes()
	if _, ok := m.dSets[DefaultDataSetName]; !ok {
		m.dSets[DefaultDataSetName] = m.newDataSet()
	}
	return m
}

// newDataSet returns a new initialize *dataSet.
func (m *Model) newDataSet() *dataSet {
	xs := float64(m.GraphWidth()) / (m.ViewMaxX() - m.ViewMinX()) // X scale factor
	ys := float64(m.Origin().Y) / (m.ViewMaxY() - m.ViewMinY())   // y scale factor
	ds := &dataSet{
		LineStyle: m.dLineStyle,
		Style:     m.dStyle,
		pBuf: buffer.NewFloat64PointScaleBuffer(
			canvas.Float64Point{X: m.ViewMinX(), Y: m.ViewMinY()},
			canvas.Float64Point{X: xs, Y: ys}),
	}
	return ds
}

// getLineSequence returns a sequence of Y values
// to draw line runes from a given set of scaled []FloatPoint64.
func (m *Model) getLineSequence(points []canvas.Float64Point) (seqY []int) {
	// Create a []int storing canvas coordinates to draw line runes.
	// Each index of the []int corresponds to a canvas column
	// and the value of each index is the canvas row
	// I.E. (X,seqY[X]) coorindates will be used to draw the line runes
	width := m.Width() - m.Origin().X // lines can draw on Y axis
	seqY = make([]int, width, width)

	// initialize every index to the value such that
	// a horizontal line at Y = 0 will be drawn
	f := m.ScaleFloat64Point(canvas.Float64Point{X: 0.0, Y: 0.0})
	for i := range seqY {
		seqY[i] = canvas.CanvasYCoordinate(m.Origin().Y, int(math.Round(f.Y)))
		// avoid drawing below X axis
		if (m.XStep() > 0) && (seqY[i] > m.Origin().Y) {
			seqY[i] = m.Origin().Y
		}
	}
	// map data set containing scaled Float64Point data points
	// onto graph row and column
	for _, p := range points {
		m.setLineSequencePoint(seqY, p)
	}
	return
}

// setLineSequencePoint will map a scaled Float64Point data point
// on to a sequence of graph Y values.  Points mapping onto
// existing indices of the sequence will override the existing value.
func (m *Model) setLineSequencePoint(seqY []int, f canvas.Float64Point) {
	x := int(math.Round(f.X))
	// avoid drawing outside graphing area
	if (x >= 0) && (x < len(seqY)) {
		// avoid drawing below X axis
		seqY[x] = canvas.CanvasYCoordinate(m.Origin().Y, int(math.Round(f.Y)))
		if (m.XStep() > 0) && (seqY[x] > m.Origin().Y) {
			seqY[x] = m.Origin().Y
		}
	}
}

// rescaleData will scale all internally stored data with new scale factor.
func (m *Model) rescaleData() {
	// rescale all data set graph points
	xs := float64(m.GraphWidth()) / (m.ViewMaxX() - m.ViewMinX()) // X scale factor
	ys := float64(m.Origin().Y) / (m.ViewMaxY() - m.ViewMinY())   // y scale factor
	offset := canvas.Float64Point{X: m.ViewMinX(), Y: m.ViewMinY()}
	scale := canvas.Float64Point{X: xs, Y: ys}
	for _, ds := range m.dSets {
		ds.pBuf.SetOffset(offset)
		ds.pBuf.SetScale(scale) // buffer rescales all raw data points
	}
}

// ClearAllData will reset stored data values in all data sets.
func (m *Model) ClearAllData() {
	for n := range m.dSets {
		m.ClearDataSet(n)
	}
}

// ClearDataSet will erase stored data set given by name string.
func (m *Model) ClearDataSet(n string) {
	if _, ok := m.dSets[n]; ok {
		delete(m.dSets, n)
	}
}

// SetXRange updates the minimum and maximum expected X values.
func (m *Model) SetXRange(min, max unit.Frequency) {
	m.Model.SetXRange(float64(min), float64(max))
}

// SetYRange updates the minimum and maximum expected Y values.
func (m *Model) SetYRange(min, max unit.Decabel) {
	m.Model.SetYRange(float64(min), float64(max))
}

// SetViewXYRange updates the displayed minimum and maximum X and Y values.
// Existing data will be rescaled.
func (m *Model) SetXYRange(minX, maxX unit.Frequency, minY, maxY unit.Decabel) {
	m.Model.SetXRange(float64(minX), float64(maxX))
	m.Model.SetYRange(float64(minY), float64(maxY))
}

// SetViewXRange updates the displayed minimum and maximum X values.
// Existing data will be rescaled.
func (m *Model) SetViewXRange(min, max unit.Frequency) {
	m.Model.SetViewXRange(float64(min), float64(max))
	m.rescaleData()
}

// SetViewYRange updates the displayed minimum and maximum Y values.
// Existing data will be rescaled.
func (m *Model) SetViewYRange(min, max unit.Decabel) {
	m.Model.SetViewYRange(float64(min), float64(max))
	m.rescaleData()
}

// SetViewXYRange updates the displayed minimum and maximum X and Y values.
// Existing data will be rescaled.
func (m *Model) SetViewXYRange(minX, maxX unit.Frequency, minY, maxY unit.Decabel) {
	m.Model.SetViewXRange(float64(minX), float64(maxX))
	m.Model.SetViewYRange(float64(minY), float64(maxY))
	m.rescaleData()
}

// Resize will change wavelinechart display width and height.
// Existing data will be rescaled.
func (m *Model) Resize(w, h int) {
	m.Model.Resize(w, h)
	m.rescaleData()
}

// SetStyles will set the default styles of data sets.
func (m *Model) SetStyles(ls runes.LineStyle, s lipgloss.Style) {
	m.dLineStyle = ls
	m.dStyle = s
	m.SetDataSetStyles(DefaultDataSetName, ls, s)
}

// SetDataSetStyles will set the styles of the given data set by name string.
func (m *Model) SetDataSetStyles(n string, ls runes.LineStyle, s lipgloss.Style) {
	if _, ok := m.dSets[n]; !ok {
		m.dSets[n] = m.newDataSet()
	}
	ds := m.dSets[n]
	ds.LineStyle = ls
	ds.Style = s
}

// Plot will map a Float64Point data value to a canvas coordinates
// to be displayed with Draw. Uses default data set.
func (m *Model) Plot(f canvas.Float64Point) {
	m.PlotDataSet(DefaultDataSetName, f)
}

// PlotDataSet will map a Float64Point data value to a canvas coordinates
// to be displayed with Draw. Uses given data set by name string.
func (m *Model) PlotDataSet(n string, f canvas.Float64Point) {
	if m.AutoAdjustRange(f) { // auto adjust x and y ranges if enabled
		m.UpdateGraphSizes()
		m.rescaleData()
	}
	if _, ok := m.dSets[n]; !ok {
		m.dSets[n] = m.newDataSet()
	}
	ds := m.dSets[n]
	ds.pBuf.Push(f)
}

// Draw will draw lines runes for each column
// of the graphing area of the canvas. Uses default data set.
func (m *Model) Draw() {
	m.DrawDataSets([]string{DefaultDataSetName})
}

// DrawAll will draw lines runes for each column
// of the graphing area of the canvas for all data sets.
// Will always draw default data set.
func (m *Model) DrawAll() {
	names := make([]string, 0, len(m.dSets))
	for n, ds := range m.dSets {
		if (n == DefaultDataSetName) || (ds.pBuf.Length() > 0) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	m.DrawDataSets(names)
}

// drawYLabel draws Y axis values left of the Y axis every n step.
// Repeating values will be hidden.
// Does nothing if n <= 0.
func (m *Model) drawYLabel(n int) {
	// from origin going up, draw data value left of the Y axis every n steps
	// origin X coordinates already set such that there is space available
	if n <= 0 {
		return
	}
	var lastVal string
	rangeSz := m.ViewMaxY() - m.ViewMinY() // range of possible expected values

	for i := range n {
		placement := float64(i) / float64(n-1)

		v := m.ViewMinY() + rangeSz*placement // value to set left of Y axis
		s := m.YLabelFormatter(i, v)

		x := m.Origin().X - len(s)
		y := m.Origin().Y - int(float64(m.GraphHeight())*placement)

		if lastVal != s {
			m.Canvas.SetStringWithStyle(canvas.Point{X: x, Y: y}, s, m.LabelStyle)
			lastVal = s
		}
		i += n
	}
}

// drawXLabel draws X axis values below the X axis every n step.
// Repeating values will be hidden.
// Does nothing if n <= 0.
func (m *Model) drawXLabel(n int) {
	// from origin going right, draw data value left of the Y axis every n steps
	if n <= 0 {
		return
	}

	var lastVal string
	rangeX := m.ViewMaxX() - m.ViewMinX()

	origin := m.Origin()

	for i := range n {
		placement := float64(i) / float64(n-1)

		v := m.ViewMinX() + rangeX*placement
		s := m.XLabelFormatter(i, v)

		align := int(float64(len(s)) * placement)

		x := origin.X + int(float64(m.GraphWidth())*placement) - align
		y := origin.Y + 1

		// can only set if rune to the left of target coordinates is empty
		if c := m.Canvas.Cell(canvas.Point{X: x - 1, Y: y}); c.Rune == runes.Null {

			// dont display if number will be cut off or value repeats
			sLen := len(s) + origin.X + i

			if (s != lastVal) && (sLen <= m.Canvas.Width()) {
				m.Canvas.SetStringWithStyle(canvas.Point{X: x, Y: y}, s, m.LabelStyle)
				lastVal = s
			}
		}
	}
}

// DrawXYAxisAndLabel draws the X, Y axes.
func (m *Model) DrawXYAxisAndLabel() {
	drawY := m.YStep() > 0
	drawX := m.XStep() > 0

	if drawY && drawX {
		graph.DrawXYAxis(&m.Canvas, m.Origin(), m.AxisStyle)
	} else {
		if drawY { // draw Y axis
			graph.DrawVerticalLineUp(&m.Canvas, m.Origin(), m.AxisStyle)
		}
		if drawX { // draw X axis
			graph.DrawHorizonalLineRight(&m.Canvas, m.Origin(), m.AxisStyle)
		}
	}

	m.drawYLabel(m.YStep())
	m.drawXLabel(m.XStep())
}

// DrawDataSets will draw lines runes for each column
// of the graphing area of the canvas for each data set given
// by name strings.
func (m *Model) DrawDataSets(names []string) {
	if len(names) == 0 {
		return
	}
	m.Clear()
	m.DrawXYAxisAndLabel()
	for _, n := range names {
		if ds, ok := m.dSets[n]; ok {
			startX := m.Origin().X
			seqY := m.getLineSequence(ds.pBuf.ReadAll())
			graph.DrawLineSequence(&m.Canvas,
				true,
				startX,
				seqY,
				ds.LineStyle,
				ds.Style)
		}
	}
}

// DrawBraille will draw braille runes displayed from left to right
// of the graphing area of the canvas. Uses default data set.
func (m *Model) DrawBraille() {
	m.DrawBrailleDataSets([]string{DefaultDataSetName})
}

// DrawBrailleAll will draw braille runes for all data sets
// from left to right of the graphing area of the canvas.
func (m *Model) DrawBrailleAll() {
	names := make([]string, 0, len(m.dSets))
	for n, ds := range m.dSets {
		if ds.pBuf.Length() > 0 {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	m.DrawBrailleDataSets(names)
}

// DrawBrailleDataSets will draw braille runes from left to right
// of the graphing area of the canvas for each data set given
// by name strings.
func (m *Model) DrawBrailleDataSets(names []string) {
	if len(names) == 0 {
		return
	}
	m.Clear()
	m.DrawXYAxisAndLabel()
	for _, n := range names {
		if ds, ok := m.dSets[n]; ok {
			dataPoints := ds.pBuf.ReadAll()
			dataLen := len(dataPoints)
			if dataLen == 0 {
				return
			}
			// draw lines from each point to the next point
			bGrid := graph.NewBrailleGrid(m.GraphWidth(), m.GraphHeight(),
				0, float64(m.GraphWidth()), // X values already scaled to graph
				0, float64(m.GraphHeight())) // Y values already scaled to graph
			for i := range dataLen {
				j := i + 1
				if j >= dataLen {
					j = i
				}
				p1 := dataPoints[i]
				p2 := dataPoints[j]
				// ignore points that will not be displayed
				bothBeforeMin := (p1.X < 0 && p2.X < 0)
				bothAfterMax := (p1.X > float64(m.GraphWidth()) && p2.X > float64(m.GraphWidth()))
				if bothBeforeMin || bothAfterMax {
					continue
				}
				// get braille grid points from two Float64Point data points
				gp1 := bGrid.GridPoint(p1)
				gp2 := bGrid.GridPoint(p2)
				// set all points in the braille grid
				// between two points that approximates a line
				points := graph.GetLinePoints(gp1, gp2)
				for _, p := range points {
					bGrid.Set(p)
				}
			}

			// get all rune patterns for braille grid
			// and draw them on to the canvas
			startX := 0
			if m.YStep() > 0 {
				startX = m.Origin().X + 1
			}
			patterns := bGrid.BraillePatterns()
			graph.DrawBraillePatterns(&m.Canvas,
				canvas.Point{X: startX, Y: 0}, patterns, ds.Style)
		}
	}
}

// Update processes bubbletea Msg to by invoking
// UpdateMsgHandlerFunc callback if wavelinechart is focused.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// if !m.Focused() {
	// 	return m, nil
	// }

	m.UpdateHandler(&m.Model, msg)
	m.rescaleData() // rescale data points to new viewing window

	return m, nil
}
