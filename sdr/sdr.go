package sdr

import (
	"fmt"
	"log"

	"github.com/olistrik/numa-sdr/api/signal"
	"github.com/olistrik/numa-sdr/api/unit"
	"github.com/pothosware/go-soapy-sdr/pkg/device"
	"github.com/pothosware/go-soapy-sdr/pkg/sdrlogger"
)

// #include <unistd.h>
// #include <fcntl.h>
// #include <stdio.h>
//
// int silence_stderr(int *oldfd) {
//     *oldfd = dup(STDERR_FILENO);
//     if (*oldfd == -1) return -1;
//     int devnull = open("/dev/null", O_WRONLY);
//     if (devnull == -1) return -1;
//     return dup2(devnull, STDERR_FILENO);
// }
//
// int restore_stderr(int oldfd) {
//     int result = dup2(oldfd, STDERR_FILENO);
//     close(oldfd);
//     return result;
// }
import "C"

func callSilently(f func()) {
	var oldfd C.int
	if C.silence_stderr(&oldfd) != 0 {
		// handle error if needed
		f()
		return
	}

	f()

	C.restore_stderr(oldfd)
}

type Sdr struct {
	device *device.SDRDevice

	samples   uint
	frequency unit.Frequency
	hops      uint

	Signal signal.WindowSignal
}

func New(args map[string]string) Sdr {
	var sdr Sdr

	callSilently(func() {
		sdrlogger.SetLogLevel(sdrlogger.Critical)
		dev, err := device.Make(args)
		if err != nil {
			log.Fatal(err)
		}

		sdr = Sdr{device: dev, samples: 512}
		sdr.SetHops(0)
		dev.SetGainMode(device.DirectionRX, 0, false)
		dev.WriteSetting("direct_samp", "0")
		dev.SetIQBalance(device.DirectionRX, 0, 1, 1)
	})

	return sdr
}

func (sdr *Sdr) SetHops(value uint) {
	sdr.hops = value
	sdr.Signal = signal.NewWindowSignal(sdr.samples, sdr.hops)
}

func (sdr Sdr) Hops() uint {
	return sdr.hops
}

func (sdr Sdr) TotalSamples() uint {
	return sdr.hops * sdr.samples
}

func (sdr Sdr) TotalBandwidth() unit.Frequency {
	return sdr.SampleRate() * unit.Frequency(sdr.hops)
}

func (sdr *Sdr) SetFreqency(frq unit.Frequency) {
	sdr.frequency = frq
}

func (sdr Sdr) Freqency() unit.Frequency {
	return sdr.frequency
}

func (sdr Sdr) SetSampleRate(rate unit.Frequency) error {
	return sdr.device.SetSampleRate(device.DirectionRX, 0, float64(rate))
}

func (sdr Sdr) SampleRate() unit.Frequency {
	return unit.Frequency(sdr.device.GetSampleRate(device.DirectionRX, 0))
}

func (sdr Sdr) SupportedSampleRates() []unit.Frequency {
	ranges := sdr.device.GetSampleRateRange(device.DirectionRX, 0)

	rates := []unit.Frequency{}
	for _, rng := range ranges {
		fmt.Println(rng)
	}

	return rates
}

func (sdr *Sdr) Close() error {
	return sdr.device.Unmake()
}

func (sdr *Sdr) Stream() (func() error, error) {
	// channel to communcate with routine.
	in := make(chan error, 1)
	out := make(chan error, 1)

	// function to kill the routine.
	stopStream := func() error {
		// any input will kill the routine.
		in <- nil

		// wait for the response.
		return <-out
	}

	go func() {
		// double wrapped routine to simplify out <- channel.
		out <- func() error {
			// // start a new stream in the routine.
			// stream, err := sdr.device.SetupSDRStreamCS8(
			// 	device.DirectionRX,
			// 	[]uint{0},
			// 	nil,
			// )
			// if err != nil {
			// 	return err
			// }

			// all is well, push nil to let main routine know.
			out <- nil

			dropBuffer := make([]int8, 2048)
			drop := make([][]int8, 1)
			drop[0] = dropBuffer

			// init buffers.
			buffer := make([]int8, 2*sdr.samples)
			buffers := make([][]int8, 1)
			buffers[0] = buffer

			flags := make([]int, 1)

			var hop uint = 0

		loop:
			for {
				select {
				case <-in:
					break loop
				default:
				}

				hop = (hop + 1) % sdr.hops

				sampleRate := sdr.SampleRate()
				// sampleTime := unit.Frequency(sdr.samples) / sampleRate * unit.Frequency(time.Second)

				sweepFrequency := sdr.frequency - sdr.TotalBandwidth()/2 + sampleRate/2 + sampleRate*unit.Frequency(hop)
				if err := sdr.device.SetFrequency(device.DirectionRX, 0, float64(sweepFrequency), nil); err != nil {
					panic(err)
				}

				// start a new stream in the routine.
				stream, err := sdr.device.SetupSDRStreamCS8(
					device.DirectionRX,
					[]uint{0},
					nil,
				)
				if err != nil {
					panic(err)
				}

				if err := stream.Activate(0, 0, 0); err != nil {
					panic(err)
				}

				_, _, err = stream.Read(drop, 2048, flags, 100000)
				if err != nil {
				}

				_, _, err = stream.Read(buffers, sdr.samples, flags, 100000)
				if err != nil {
				}

				if err := stream.Deactivate(0, 0); err != nil {
					panic(err)
				}

				if err := stream.Close(); err != nil {
					fmt.Printf("Error closing stream: %v\n", err)
				}

				for idx := range sdr.samples {
					br := idx * 2
					bi := br + 1

					sdr.Signal[hop][idx] = complex(float64(buffer[br]), float64(buffer[bi]))
				}
			}

			return nil
		}()
	}()

	// block until setup is successful or errors.
	if err := <-out; err != nil {
		return nil, err
	}

	return stopStream, nil
}
