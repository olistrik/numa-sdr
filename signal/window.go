package signal

import (
	"slices"
)

type WindowSignal []Signal

func NewWindowSignal(samples uint, bins uint) WindowSignal {
	signal := make([]Signal, bins)

	for i := range signal {
		signal[i] = make(Signal, samples)
	}

	return signal
}

func (ws WindowSignal) Signal() Signal {
	return slices.Concat(ws...)
}

func (ws WindowSignal) Fft() Signal {
	signal := make(Signal, len(ws.Signal()))

	for i, bin := range ws {
		offset := i * len(bin)
		copy(signal[offset:], bin.Fft().Shifted())
	}

	return signal
}
