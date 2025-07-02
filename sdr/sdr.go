package sdr

import (
	"fmt"
	"log"

	"github.com/olistrik/numa-sdr/api/signal"
	"github.com/olistrik/numa-sdr/api/unit"
	"github.com/pothosware/go-soapy-sdr/pkg/device"
)

type Sdr struct {
	samples uint
	signal.Signal

	device *device.SDRDevice

	frequency  unit.Frequency
	sweepSteps uint
}

func New(args map[string]string) Sdr {
	dev, err := device.Make(args)
	if err != nil {
		log.Fatal(err)
	}

	sdr := Sdr{device: dev, samples: 512}
	sdr.SetSweepSteps(0)
	dev.SetGainMode(device.DirectionRX, 0, false)
	dev.WriteSetting("direct_samp", "0")
	dev.SetIQBalance(device.DirectionRX, 0, 1, 1)
	// fmt.Println(dev.GetIQBalance(device.DirectionRX, 0))

	return sdr
}

func (sdr *Sdr) SetSweepSteps(value uint) {
	sdr.sweepSteps = value
	sdr.Signal = signal.New(sdr.SweepSamples())
}

func (sdr Sdr) SweepSteps() uint {
	return sdr.sweepSteps
}

func (sdr Sdr) SweepSamples() uint {
	return (sdr.sweepSteps + 1) * sdr.samples
}

func (sdr Sdr) SweepWidth() unit.Frequency {
	return sdr.SampleRate() * unit.Frequency(sdr.sweepSteps+1)
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
			fmt.Print("pushing")
			// start a new stream in the routine.
			stream, err := sdr.device.SetupSDRStreamCS8(
				device.DirectionRX,
				[]uint{0},
				nil,
			)
			if err != nil {
				return err
			}

			defer func() {
				if err := stream.Close(); err != nil {
					fmt.Printf("Error closing stream: %v\n", err)
				}
			}()

			// all is well, push nil to let main routine know.
			out <- nil

			// init buffers.
			buffer := make([]int8, 2*sdr.samples)
			buffers := make([][]int8, 1)
			buffers[0] = buffer

			flags := make([]int, 1)

			var sweepStep uint = 0

			if err := stream.Activate(0, 0, 0); err != nil {
				panic(err)
			}

		loop:
			for {
				select {
				case <-in:
					break loop
				default:
				}

				sweepStep = (sweepStep + 1) % (sdr.sweepSteps + 1)

				offset := 2 * sdr.samples * sweepStep

				sampleRate := sdr.SampleRate()
				// sampleTime := unit.Frequency(sdr.samples) / sampleRate * unit.Frequency(time.Second)

				sweepFrequency := sdr.frequency - sdr.SweepWidth()/2 + sampleRate/2 + sampleRate*unit.Frequency(sweepStep)
				if err := sdr.device.SetFrequency(device.DirectionRX, 0, float64(sweepFrequency), nil); err != nil {
					panic(err)
				}

				// if err := stream.Activate(0, 0, 0); err != nil {
				// 	panic(err)
				// }
				//
				_, _, err := stream.Read(buffers, sdr.samples, flags, 100000)
				if err != nil {
				}
				// if err := stream.Deactivate(0, 0); err != nil {
				// 	panic(err)
				// }

				for i := range buffer {
					sdr.Signal[i+int(offset)] = buffer[i]
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
