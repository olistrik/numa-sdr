package signal

import (
	"math"
	"math/cmplx"

	"github.com/scientificgo/fft"
)

type Signal []complex128

func NewSignal(samples uint) Signal {
	buffer := make([]complex128, samples)
	for i := range buffer {
		buffer[i] = 0
	}

	return buffer
}

func (s Signal) Raw() []complex128 {
	return s
}

func (s Signal) Hann() Signal {
	n := len(s)
	hann := make([]complex128, n)

	for i, v := range s {
		hann[i] = v * complex(0.5-0.5*math.Cos((2.*math.Pi*float64(n))/float64(n-1)), 0)
	}

	return hann
}

func (s Signal) Fft() Signal {
	return fft.Fft(s, false)
}

func (s Signal) Shifted() Signal {
	n := len(s)
	half := n / 2
	shifted := make([]complex128, n)

	copy(shifted[:n-half], s[half:])
	copy(shifted[n-half:], s[:half])

	return shifted
}

func (s Signal) Decabels() []float64 {
	result := make([]float64, len(s))

	for i, val := range s {
		result[i] = 10 * math.Log10(math.Pow(cmplx.Abs(val), 2))
	}

	return result
}

func (s Signal) Size() int {
	return len(s)
}
