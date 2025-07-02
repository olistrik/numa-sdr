package signal

import (
	"math"
	"math/cmplx"

	"github.com/scientificgo/fft"
)

type Signal []int8

func New(samples uint) Signal {
	buffer := make([]int8, 2*samples)
	for i := range buffer {
		buffer[i] = 0
	}

	return buffer
}

func (s Signal) Complex() []complex128 {
	sz := len(s) / 2
	result := make([]complex128, sz)

	for i := range result {
		br := i * 2
		bi := br + 1
		result[i] = complex(float64(s[br]), float64(s[bi]))
	}

	return result
}

func (s Signal) Fft() []complex128 {
	return fft.Fft(s.Complex(), false)
}

func (s Signal) FftShifted() []complex128 {
	fft := s.Fft()
	n := len(fft)
	half := n / 2
	shifted := make([]complex128, n)
	copy(shifted[:n-half], fft[half:])
	copy(shifted[n-half:], fft[:half])
	return shifted
}

func (s Signal) Decabels() []float64 {
	fft := s.FftShifted()

	result := make([]float64, len(fft))

	for i, val := range fft {
		result[i] = 10 * math.Log10(math.Pow(cmplx.Abs(val), 2))
	}

	return result
}
