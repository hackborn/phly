package phly_nodes_test

import (
	"errors"
	"flag"
	"fmt"
	"github.com/hackborn/phly"
	"github.com/hackborn/phly/nodes"
	"github.com/micro-go/lock"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	modePtr := flag.String("mode", "", "test mode")
	echoPtr := flag.String("echo", "", "echo value")
	limitPtr := flag.Int("limit", -1, "maximum number of counts to print")
	flag.Parse()

	switch *modePtr {
	case "":
		// Normal test mode
		os.Exit(m.Run())

	case "echo":
		// Echo back the argument and stop
		fmt.Println(*echoPtr)

	case "count":
		// Endlessly count
		i := 0
		for {
			if *limitPtr < 0 || i < *limitPtr {
				fmt.Println(strconv.Itoa(i))
			}
			time.Sleep(10 * time.Millisecond)
			i++
		}
	}
}

// ----------------------------------------
// RUN-NODE

func TestRunNode(t *testing.T) {
	cases := []struct {
		StartPins      phly.Pins
		WantStartErr   error
		WantOutputPins phly.Pins
		StopCond       nodeOutputCond
		WantStopErr    error
		Wait           time.Duration
	}{
		// Command automatically stops.
		{runEchoStartPins("hello?"), nil, phly.MustBuildPins(run_output, "hello?"), nil, nil, 0},
		// Command runs until told to stop.
		{runCountStartPins(-1), nil, phly.MustBuildPins(phly.PbsChan, run_output, "0", phly.PbsDoc, "1", phly.PbsDoc, "2"), sendPinsCond(3), err_exit1, 0},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			args := phly.ProcessArgs{}
			n := phly_nodes.New_run()
			have_output := &testNodeOutput{}
			have_output.cond = tc.StopCond
			if have_output.cond == nil {
				have_output.cond = sendPinsCond(-1)
			}
			have_start_err := n.Process(args, phly.NodeStarting, tc.StartPins, have_output)
			if have_start_err != tc.WantStartErr {
				fmt.Println("start err mismatch\nhave\n", have_start_err, "\nwant\n", tc.WantStartErr)
				t.Fatal()
			}
			if have_start_err == nil {
			}

			// Wait for the run to stop. Two things can stop it:
			// 1. I hit a stop condition (by receiving an explicit stop command, or matching a test stop condition)
			// 2. I timed out.
			waitForNode(have_output.cond.StopChannel(), tc.Wait)

			have_stop_err := n.StopNode(phly.StoppedArgs{})
			if !phly.ErrorsEqual(have_stop_err, tc.WantStopErr) {
				fmt.Println("stop err mismatch\nhave\n", have_stop_err, "\nwant\n", tc.WantStopErr)
				t.Fatal()
			}
			have_pins := have_output.builder.Pins()
			if !phly.StringPinsEqual(have_pins, tc.WantOutputPins) {
				fmt.Println("pins mismatch\nhave\n", phly.StringPinsToJson(have_pins), "\nwant\n", phly.StringPinsToJson(tc.WantOutputPins))
				t.Fatal()
			}
		})
	}
}

func waitForNode(stop <-chan struct{}, wait time.Duration) {
	const defaultWait = 100 * time.Millisecond
	if !(wait > 0) {
		wait = defaultWait
	}

	for {
		select {
		case <-stop:
			return
		case <-time.After(wait):
			return
		}
	}
}

// runEchoStartPins() creates new start pins configured to run echo.
func runEchoStartPins(echo string) phly.Pins {
	return runStartPins(os.Args[0], "-mode=echo", "-echo="+echo)
}

// runCountStartPins() creates new start pins configured to run count.
func runCountStartPins(limit int) phly.Pins {
	args := []string{"-mode=count"}
	if limit >= 0 {
		args = append(args, "-limit="+strconv.Itoa(limit))
	}
	return runStartPins(os.Args[0], args...)
}

func runStartPins(cmd string, args ...string) phly.Pins {
	b := phly.PinBuilder{}.Add(run_cmdinput, phly.NewStringDoc(cmd))
	doc := phly.NewStringDoc(args...)
	if len(doc.AllItems()) > 0 {
		b = b.Add(run_clainput, doc)
	}
	return b.Pins()
}

// ----------------------------------------
// NODE-OUTPUT

type testNodeOutput struct {
	cond    nodeOutputCond
	mutex   sync.Mutex
	builder phly.PinBuilder
}

func (t *testNodeOutput) SendPins(pins phly.Pins) {
	t.SendMsg(phly.MsgFromPins(pins))
}

func (t *testNodeOutput) SendMsg(msg phly.Msg) {
	t.sendMsgInternal(msg)
	t.cond.SendMsg(msg)
}

func (t *testNodeOutput) sendMsgInternal(msg phly.Msg) {
	defer lock.Locker(&t.mutex).Unlock()
	switch msg.What {
	case phly.WhatPins:
		if pins, ok := msg.Payload.(phly.Pins); ok && pins != nil {
			pins.WalkPins(func(name string, docs phly.Docs) {
				for _, d := range docs.Docs {
					t.builder = t.builder.Add(name, d)
				}
			})
		}
		//	case phly.WhatStop:
	}
}

// ----------------------------------------
// NODE-OUTPUT-COND

type nodeOutputCond interface {
	StopChannel() <-chan struct{}
	SendMsg(phly.Msg)
}

type nodeOutputSendPinsCond struct {
	mutex    sync.Mutex
	stopchan chan struct{}
	maxsends int
	cursends int
}

func sendPinsCond(max int) nodeOutputCond {
	stopchan := make(chan struct{})
	return &nodeOutputSendPinsCond{stopchan: stopchan, maxsends: max}
}

func (c *nodeOutputSendPinsCond) StopChannel() <-chan struct{} {
	return c.stopchan
}

func (c *nodeOutputSendPinsCond) SendMsg(msg phly.Msg) {
	switch msg.What {
	case phly.WhatPins:
		if c.incSends() {
			c.stop()
		}
	case phly.WhatStop:
		c.stop()
	}
}

// incSends() increments the sends, answering true if I need to stop.
func (c *nodeOutputSendPinsCond) incSends() bool {
	if c.maxsends <= 0 {
		return false
	}
	defer lock.Locker(&c.mutex).Unlock()
	c.cursends++
	return c.cursends >= c.maxsends
}

func (c *nodeOutputSendPinsCond) stop() {
	defer lock.Locker(&c.mutex).Unlock()
	if c.stopchan != nil {
		close(c.stopchan)
		c.stopchan = nil
	}
}

// ----------------------------------------
// ARGS

type testStopArgs struct {
}

// ----------------------------------------
// CONST and VAR

var (
	err_exit1 = errors.New("exit status 1") // Mimic a killed process

	// Convenience -- map private constants to the same names.
	run_cmdinput  = phly_nodes.Run_cmdinput
	run_clainput  = phly_nodes.Run_clainput
	run_output    = phly_nodes.Run_output
	run_erroutput = phly_nodes.Run_erroutput
)
