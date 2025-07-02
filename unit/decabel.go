package unit

import "fmt"

type Decabel float64

func (f Decabel) Unit() string {
	return "dB"
}

func (f Decabel) float64() float64 {
	return float64(f)
}

func (f Decabel) String() string {
	return fmt.Sprintf("%0.2f%s", f, f.Unit())
}
