package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/olistrik/numa-sdr/api/sdr"
	"github.com/olistrik/numa-sdr/api/tui/dashboard"
	"github.com/olistrik/numa-sdr/api/unit"
)

var args struct {
	Frequency  unit.Frequency `arg:"-f" default:"100e6" placeholder:"float"`
	SampleRate unit.Frequency `arg:"-r" default:"1e6" placeholder:"float"`
	SweepSteps uint           `arg:"-s" default:"0" placeholder:"uint"`
	Dump       string         `arg:"--dump" placeholder:"out.csv"`
	NoDash     bool           `arg:"--no-dashboard"`
}

func main() {
	arg.MustParse(&args)

	// Init the SDR.
	source := sdr.New(map[string]string{
		"driver": "rtlsdr",
	})
	defer source.Close()

	// Init defaults.
	source.SetFreqency(args.Frequency)
	source.SetSampleRate(args.SampleRate)
	source.SetSweepSteps(args.SweepSteps)

	close, err := source.Stream()
	if err != nil {
		log.Fatalln(err)
	}
	defer close()

	if !args.NoDash {
		dash := dashboard.New(source)
		if _, err := tea.NewProgram(dash, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	} else {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		for {
			select {
			case <-quit:
				return
			default:
			}
		}
	}
}
