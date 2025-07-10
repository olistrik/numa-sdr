package history

import (
	"fmt"
	"time"

	"github.com/olistrik/numa-sdr/api/sdr/power"
	log "github.com/sirupsen/logrus"
)

type HistoryOption func(*History)

func MaxDuration(duration time.Duration) HistoryOption {
	return func(h *History) {
		h.MaxDuration = duration
	}
}

type History struct {
	head *power.Scan
	tail *power.Scan
	next *power.Scan

	Hop          uint
	ExpectedHops uint
	MaxDuration  time.Duration `json:"max_duration"`
	Scans        []*power.Scan `json:"scans"`
}

func New(opts ...HistoryOption) *History {
	history := &History{
		MaxDuration: 0,
		Scans:       []*power.Scan{},
	}

	for _, opt := range opts {
		opt(history)
	}

	return history
}

func (hm *History) Tail() *power.Scan {
	return hm.tail
}

func (hm *History) Head() *power.Scan {
	return hm.head
}

func (hm *History) Push(scan *power.Scan) (bool, error) {
	// increment Hop; it should never 0.
	hm.Hop++

	if hm.head == nil {
		// First sweep.

		if hm.next == nil {
			// first hop. Init ExpectedHops.
			hm.ExpectedHops = 1
			hm.next = scan
			return false, nil
		}

		if hm.next.DateTime == scan.DateTime {
			// still first sweep, append.
			hm.ExpectedHops++
			return false, hm.next.Append(scan)
		}

		// hm.next is complete, scan is the start of the next sweep.
		// append next to scan, and point head at it.

		hm.Scans = append(hm.Scans, hm.next)
		hm.head = hm.next

		// setup next scan.
		hm.Hop = 1
		hm.next = scan

		// Technically this is wrong. It was completed _last_ call.
		// However, it was _this_ loop that it was appended to scan.
		return true, nil
	}

	// Sweep 2+.

	// Check if it's the start of a new scan.
	if hm.next == nil {
		hm.Hop = 1
		hm.next = scan

		return false, nil
	}

	// Check that we've not overrun on the hops.
	if hm.next.DateTime == scan.DateTime && hm.Hop > hm.ExpectedHops {
		hm.next = nil
		return false, fmt.Errorf("to many scans recieved for timestamp %s", scan.DateTime)
	}

	// Check that we've not underrun on the hops.
	if hm.next.DateTime != scan.DateTime && hm.Hop <= hm.ExpectedHops {
		hm.next = nil
		return false, fmt.Errorf("too few scans recieved for timestamp %s", scan.DateTime)
	}

	// Append it.
	if err := hm.next.Append(scan); err != nil {
		log.Infoln(hm.Hop, hm.ExpectedHops)
		log.Infoln(hm.next.DateTime, scan.DateTime)
		hm.next = nil
		return false, err
	}

	// otherwise, we only return if we've yet to reach the last hop.
	if hm.Hop < hm.ExpectedHops {
		return false, nil
	}

	// sweep complete. Append it to scans and shift the head.
	hm.Scans = append(hm.Scans, hm.next)
	hm.head = hm.next

	// start new scans.
	hm.next = nil

	// Drop scans older than MaxHistory
	if hm.MaxDuration > 0 {
		for hm.head.DateTime.Sub(hm.Scans[0].DateTime) > hm.MaxDuration {
			hm.Scans = hm.Scans[1:]
		}
	}

	hm.tail = hm.Scans[0]

	return true, nil
}
