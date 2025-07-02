package unit

import (
	"strconv"
)

type Frequency float64

const (
	Hz  Frequency = 1
	KHz           = 1000 * Hz
	MHz           = 1000 * KHz
	GHz           = 1000 * MHz
	THz           = 1000 * GHz
	PHz           = 1000 * THz
	EHz           = 1000 * PHz
	ZHz           = 1000 * EHz
)

func (f Frequency) Unit() string {
	return "Hz"
}

func (f Frequency) float64() float64 {
	return float64(f)
}

func (f Frequency) String() string {
	return FormatSI(f)
}

func (n *Frequency) UnmarshalText(b []byte) error {
	s := string(b)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*n = Frequency(val)
	return nil
}
