package dashboard

import (
	"math"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/olistrik/numa-sdr/api/sdr"
	"github.com/olistrik/numa-sdr/api/tui/chart/spectrum"
	"github.com/olistrik/numa-sdr/api/unit"
)

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return t
	})
}

type Dashboard struct {
	chart       spectrum.Model
	zoneManager *zone.Manager
	signal      *sdr.Sdr

	average *[]float64
	history *[][]float64
}

func New(source sdr.Sdr) Dashboard {
	width, height := 80, 20
	chart := spectrum.New(width, height)

	zoneManager := zone.New()
	chart.SetZoneManager(zoneManager)
	chart.Focus()

	average := make([]float64, source.SweepSamples())
	history := make([][]float64, 10)
	for i := range history {
		zeros := make([]float64, len(average))
		for i := range zeros {
			zeros[i] = 0
		}
		history[i] = zeros
	}

	return Dashboard{chart, zoneManager, &source, &average, &history}
}

func (m Dashboard) Init() tea.Cmd {

	return tick()
}

func (m Dashboard) Scan() tea.Cmd {
	m.chart.ClearAllData()

	dbs := m.signal.Decabels()
	bdw := m.signal.SweepWidth()
	step := bdw / unit.Frequency(len(dbs))
	frq := m.signal.Freqency()
	x := frq - bdw/2

	var last []float64
	last, (*m.history) = (*m.history)[0], (*m.history)[1:]
	(*m.history) = append(*m.history, last)

	for i, db := range dbs {
		y := db / float64(len((*m.history)))

		if math.IsNaN(y) || math.IsInf(y, 0) {
			y = 0
		}

		(*m.average)[i] -= last[i]
		(*m.average)[i] += y

		last[i] = y

		m.chart.Plot(canvas.Float64Point{X: float64(x), Y: (*m.average)[i]})
		// m.chart.Plot(canvas.Float64Point{X: float64(x), Y: y})

		x += step
	}

	m.chart.SetXRange(frq-bdw, frq+bdw)
	m.chart.SetYRange(0, 100)
	m.chart.SetViewXYRange(frq-bdw/2, frq+bdw/2, 0, 99)
	m.chart.DrawBrailleAll()

	return tick()
}

func (m Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.chart.Resize(msg.Width-2, msg.Height-2)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case time.Time:
		return m, m.Scan()

	}

	m.chart, _ = m.chart.Update(msg)

	return m, nil
}

func (m Dashboard) View() string {
	return m.zoneManager.Scan(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Render(m.chart.View()),
	)
}
