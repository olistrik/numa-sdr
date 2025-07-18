// ntcharts - Copyright (c) 2024 Neomantra Corp.

package spectrum

import (
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/olistrik/numa-sdr/api/unit"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// Option is used to set options when initializing a wavelinechart. Example:
//
//	wlc := New(width, height, WithStyles(someLineStyle, someLipglossStyle))
type Option func(*Model)

// WithLineChart sets internal linechart to given linechart.
func WithLineChart(lc *linechart.Model) Option {
	return func(m *Model) {
		m.Model = *lc
	}
}

// WithUpdateHandler sets the UpdateHandler used
// when processing bubbletea Msg events in Update().
func WithUpdateHandler(h linechart.UpdateHandler) Option {
	return func(m *Model) {
		m.UpdateHandler = h
	}
}

// WithZoneManager sets the bubblezone Manager used
// when processing bubbletea Msg mouse events in Update().
func WithZoneManager(zm *zone.Manager) Option {
	return func(m *Model) {
		m.SetZoneManager(zm)
	}
}

// WithXYSteps sets the number of steps when drawing X and Y axes values.
// If X steps 0, then X axis will be hidden.
// If Y steps 0, then Y axis will be hidden.
func WithXYSteps(x, y int) Option {
	return func(m *Model) {
		m.SetXStep(x)
		m.SetYStep(y)
	}
}

// WithXRange sets expected and displayed
// minimum and maximum Y value range.
func WithXRange(min, max unit.Frequency) Option {
	return func(m *Model) {
		m.SetXRange(min, max)
		m.SetViewXRange(min, max)
	}
}

// WithYRange sets expected and displayed
// minimum and maximum Y value range.
func WithYRange(min, max unit.Decabel) Option {
	return func(m *Model) {
		m.SetYRange(min, max)
		m.SetViewYRange(min, max)
	}
}

// WithXYRange sets expected and displayed
// minimum and maximum Y value range.
func WithXYRange(minX, maxX unit.Frequency, minY, maxY unit.Decabel) Option {
	return func(m *Model) {
		m.SetXRange(minX, maxX)
		m.SetViewXRange(minX, maxX)
		m.SetYRange(minY, maxY)
		m.SetViewYRange(minY, maxY)
	}
}

// WithStyles sets the default line style and lipgloss style of data sets.
func WithStyles(ls runes.LineStyle, s lipgloss.Style) Option {
	return func(m *Model) {
		m.SetStyles(ls, s)
	}
}

// WithAxesStyles sets the axes line and line label styles.
func WithAxesStyles(as lipgloss.Style, ls lipgloss.Style) Option {
	return func(m *Model) {
		m.AxisStyle = as
		m.LabelStyle = ls
	}
}

// WithDataSetStyles sets the line style and lipgloss style
// of the data set given by name.
func WithDataSetStyles(n string, ls runes.LineStyle, s lipgloss.Style) Option {
	return func(m *Model) {
		m.SetDataSetStyles(n, ls, s)
	}
}

// WithPoints maps []Float64Point data points to canvas coordinates
// for the default data set.
func WithPoints(f []canvas.Float64Point) Option {
	return func(m *Model) {
		for _, v := range f {
			m.Plot(v)
		}
	}
}

// WithDataSetPoints maps []Float64Point data points to canvas coordinates
// for the data set given by name.
func WithDataSetPoints(n string, f []canvas.Float64Point) Option {
	return func(m *Model) {
		for _, v := range f {
			m.PlotDataSet(n, v)
		}
	}
}
