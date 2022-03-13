package main

import (
	"fmt"
	"os"
	"os/signal"

	ssh "github.com/helloyi/go-sshclient"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"

	testdrv "gitlab.com/gomidi/midi/testdrv"
	midicat "gitlab.com/gomidi/midicatdrv"
	// rtmididrv "gitlab.com/gomidi/rtmididrv"
	// portmididrv "gitlab.com/gomidi/portmididrv"
	// webmididrv "gitlab.com/gomidi/webmididrv"
)

// TODO Replace all vars/consts with cobra CLI flags

// InitSequence determines the sequence that will be played on init
var InitSequence [][]int = TripleBeepFunky

const (
	// PlayInit determines whether InitSequence will be played
	PlayInit = true
	// BeepDuration determines the max duration each beep plays for in seconds
	BeepDuration = 10
	// BendDivision determines the total diference in bend before we send another command
	BendDivision = 1024
	// BendZero determines the pitchbend zero position
	BendZero = 8192
	// Driver determines the MIDI driver we're using
	Driver = Midicat
	// Host determines the SSH connection address
	Host = "192.168.88.1:22"
	// User determines the SSH username
	User = "admin"
	// Pass determines the SSH password
	Pass = ""
)

// Driver defines a midi driver name
type MidiDriver string

const (
	Test     MidiDriver = "test"
	Midicat  MidiDriver = "midicat"
	RTMidi   MidiDriver = "rtmidi"
	PortMidi MidiDriver = "portmididrv"
	WebMidi  MidiDriver = "webmididrv"
)

// MessageKey defines the keys of each midi message
type MessageKey string

const (
	Key       MessageKey = "key"
	Channel   MessageKey = "channel"
	Velocity  MessageKey = "velocity"
	Frequency MessageKey = "frequency"
)

// Action defines a Midi action
type Action int

const (
	Invalid Action = iota
	NoteOff
	NoteOn
	Pitchbend
)

// Midi defines a midi message
type Midi struct {
	action   Action
	channel  int
	key      int
	velocity int
	value    int
}

func main() {
	fmt.Printf("Connecting to ssh: %s\n", Host)

	client, err := ssh.DialWithPasswd(Host, User, Pass)
	if err != nil {
		must(err)
	}
	defer func() {
		fmt.Println("Closing ssh")
		client.Close()
	}()

	fmt.Println("Connection successful!")

	mikrotikBeeper, err := NewMikroTikBeeper(client)
	must(err)

	limitedBeeper, err := NewLimitedBeeper(mikrotikBeeper, BendDivision)
	must(err)

	beeper, err := NewMidiBeeper(limitedBeeper)
	must(err)

	defer func() {
		fmt.Println("Closing beeper")
		beeper.NoBeep(-1)
	}()

	fmt.Printf("Connecting to midi using: %s\n", Driver)

	var drv midi.Driver
	switch Driver {
	case Test:
		drv = testdrv.New("MidiTik")
	case Midicat:
		drv, err = midicat.New()
	// case RTMidi:
	// 	drv, err = rtmididrv.New()
	// case PortMidi:
	// 	drv, err = portmididrv.New()
	// case WebMidi:
	// 	drv, err = webmididrv.New()
	default:
		err = fmt.Errorf("invalid midi driver")
	}
	must(err)

	defer func() {
		fmt.Println("Closing midi driver")
		drv.Close()
	}()

	ins, err := drv.Ins()
	must(err)
	outs, err := drv.Outs()
	must(err)

	in, out := ins[0], outs[0]
	must(in.Open())
	must(out.Open())

	defer func() {
		fmt.Println("Closing input")
		in.Close()
	}()
	defer func() {
		fmt.Println("Closing output")
		out.Close()
	}()

	bend := BendZero

	rd := reader.New(
		reader.NoLogger(),
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			midi, err := decodeMidi(msg)
			if err != nil {
				fmt.Printf("failed to deocde midi: %v\n", err)
				return
			}

			switch midi.action {
			case NoteOn:
				err = beeper.Beep(midi.key, bend)
			case NoteOff:
				err = beeper.NoBeep(midi.key)
			case Pitchbend:
				bend = midi.value
				err = beeper.Beep(midi.key, bend)
			default:
				err = fmt.Errorf("invalid action: %v", midi.action)
			}
			if err != nil {
				fmt.Printf("Could not process midi: %v\n", err)
			}
		}),
	)

	err = rd.ListenTo(in)
	must(err)

	if PlayInit {
		playSequence(beeper, InitSequence[0], InitSequence[1])
		// randomBeeps(beeper, midiMap)
	}

	fmt.Println("Connection successful!")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
}
