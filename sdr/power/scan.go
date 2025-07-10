package power

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olistrik/numa-sdr/api/unit"
)

type Scan struct {
	DateTime       time.Time      `json:"date_time"`
	StartFrequency unit.Frequency `json:"start_frequency"`
	EndFrequency   unit.Frequency `json:"end_frequency"`
	SampleRate     unit.Frequency `json:"sample_rate"`
	SampleCount    uint           `json:"-"` // This is kind of useless.
	Bins           []unit.Decabel `json:"bins"`
}

func ParseScan(line string) (*Scan, error) {
	scan := &Scan{}

	data := strings.Split(line, ",")
	for i := range data {
		data[i] = strings.Trim(data[i], " ")
	}

	/* DateTime */
	date, err := time.Parse(time.DateTime, data[0]+" "+data[1])
	if err != nil {
		return nil, err
	}
	scan.DateTime = date

	/*  StartFrequency */
	startFrequency, err := strconv.ParseFloat(data[2], 64)
	if err != nil {
		return nil, err
	}
	scan.StartFrequency = unit.Frequency(startFrequency)

	/* EndFrequency */
	endFrequency, err := strconv.ParseFloat(data[3], 64)
	if err != nil {
		return nil, err
	}
	scan.EndFrequency = unit.Frequency(endFrequency)

	/* SampleRate */
	sampleRate, err := strconv.ParseFloat(data[4], 64)
	if err != nil {
		return nil, err
	}
	scan.SampleRate = unit.Frequency(sampleRate)

	/* SampleCount */
	sampleCount, err := strconv.ParseUint(data[5], 10, 0)
	if err != nil {
		return nil, err
	}
	scan.SampleCount = uint(sampleCount)

	bins := data[6:]

	scan.Bins = make([]unit.Decabel, len(bins))

	/* Bins */
	for i, bin := range bins {
		value, err := strconv.ParseFloat(bin, 64)
		if err != nil {
			return nil, err
		}
		scan.Bins[i] = unit.Decabel(value)
	}

	return scan, nil
}

func (scan *Scan) Append(next *Scan) error {
	if scan.DateTime != next.DateTime {
		return fmt.Errorf("refusing to append scans with different datetimes")
	}

	if scan.SampleRate != next.SampleRate {
		return fmt.Errorf("refusing to append scans with different samplerates")
	}

	if next.StartFrequency < scan.EndFrequency {
		return fmt.Errorf("refusing to append non-sequentual scans")
	}

	scan.EndFrequency = next.EndFrequency
	scan.Bins = append(scan.Bins, next.Bins...)

	return nil
}
