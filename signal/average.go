package signal

import "time"

type SignalLike interface {
	Fft() Signal
	Size() int
}

type timeSignal struct {
	time   time.Time
	signal Signal
}

type TimeAverager struct {
	source   SignalLike
	duration time.Duration

	history []timeSignal
	average Signal
}

func NewAverager(signal SignalLike, window uint) TimeAverager {

	return TimeAverager{
		source: signal,

		average: make(Signal, signal.Size()),
		history: []timeSignal{},
	}
}

func (a *TimeAverager) Update() {
	now := time.Now()
	old := 0

	for i, val := range a.history {
		expiry := val.time.Add(a.duration)
		if now.Before(expiry) {
			break
		}
		old = i
	}

	_ = old
}
