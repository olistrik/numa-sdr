package unit

import (
	"fmt"
	"math"
)

type Unit interface {
	Unit() string
	float64() float64
}

const muliplePrefixes = "_kMGTPEZY"
const submultiplePrefixes = "_mµnpfazy"

func FormatSI(value Unit) string {
	power := int(math.Floor(math.Log10(math.Abs(value.float64()))))

	idx := power / 3

	prefix := ""

	if idx > 0 {
		power = min(idx, len(muliplePrefixes)-1)
		prefix = string(muliplePrefixes[power])
	}

	if -idx >= len(submultiplePrefixes) {
		idx = 0
	}

	if idx < 0 {
		power = min(-idx, len(submultiplePrefixes)-1)
		prefix = string(submultiplePrefixes[power])
	}

	scaled := value.float64() / math.Pow10(idx*3)

	return fmt.Sprintf("%0.2f%s%s", scaled, prefix, value.Unit())
}
