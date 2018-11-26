package phly_nodes

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hackborn/phly"
	"github.com/micro-go/lock"
	"github.com/micro-go/parse"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// run executes a command.
type run struct {
	runner *run_func_t
}

func (n *run) Describe() phly.NodeDescr {
	descr := phly.NodeDescr{Id: "phly/run", Name: "Run", Purpose: "Run a program."}
	descr.StartupPins = append(descr.StartupPins, phly.PinDescr{Name: run_cmdinput, Purpose: "The command to run."})
	descr.StartupPins = append(descr.StartupPins, phly.PinDescr{Name: run_clainput, Purpose: "Optional command line arguments."})
	descr.OutputPins = append(descr.OutputPins, phly.PinDescr{Name: run_output, Purpose: "Standard output from the running command."})
	descr.OutputPins = append(descr.OutputPins, phly.PinDescr{Name: run_erroutput, Purpose: "Error output from the running command."})
	return descr
}

func (n *run) Instantiate(args phly.InstantiateArgs, cfg interface{}) (phly.Node, error) {
	return &run{}, nil
}

func (n *run) Process(args phly.ProcessArgs, stage phly.NodeStage, input phly.Pins, output phly.NodeOutput) error {
	if stage == phly.NodeStarting {
		err := n.startNode(args, input, output)
		if err != nil {
			return err
		}
	}
	// XXX handle input
	return nil
}

func (n *run) startNode(args phly.ProcessArgs, input phly.Pins, output phly.NodeOutput) error {
	// Swallow close errors, which should have been handled in the stop.
	n.close()

	cmd := input.GetPin(run_cmdinput).StringItem(0)
	cla := input.GetPin(run_clainput).StringItems()
	// Validate
	if cmd == "" {
		return phly.BadRequestErr
	}
	// Support two formats -- either passing in an array of CLAs, or a single string that needs to be parsed.
	if len(cla) == 1 {
		cla = parse.AsArguments(cla[0])
	}

	n.runner = startRunFunc(args, output, cmd, cla)

	return nil
}

func (n *run) StopNode(args phly.StoppedArgs) error {
	// Answer any error generated during stopping
	err, _ := n.close()
	return err
}

// close() closes my current runner, answering both the error
// returned from the run process, and any error that occurred trying to close.
func (n *run) close() (error, error) {
	if n.runner != nil {
		erra, errb := n.runner.close()
		n.runner = nil
		return erra, errb
	}
	return nil, nil
}

// ----------------------------------------
// RUN-FUNC-T

// run_func_t struct is the state of a currently running node.
type run_func_t struct {
	cmd  *exec.Cmd
	wait *sync.WaitGroup
	err  lock.AtomicError // Store the current state of the running operation, or its result
}

func startRunFunc(args phly.ProcessArgs, output phly.NodeOutput, _cmd string, _cla []string) *run_func_t {
	cmd := exec.Command(args.Filename(_cmd), _cla...)
	fn := &run_func_t{cmd: cmd, wait: &sync.WaitGroup{}, err: lock.NewAtomicError()}
	fn.err.SetTo(run_node_starting)
	fn.wait.Add(1)
	go fn.run(cmd, output)
	return fn
}

func (r *run_func_t) run(cmd *exec.Cmd, output phly.NodeOutput) {
	var err error
	var stdout io.ReadCloser
	var stderr io.ReadCloser

	defer r.wait.Done()
	defer func() { r.err.SetTo(err) }()
	defer func() { output.SendMsg(phly.MsgFromStop(err)) }()

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	go func() {
		streamOutput(stdout, run_output, output)
	}()
	go func() {
		streamOutput(stderr, run_erroutput, output)
	}()

	err = cmd.Start()
	if err != nil {
		return
	}

	r.err.SetTo(run_node_running)
	err = cmd.Wait()
}

// close() closes the runner, answering both the error returned
// from the run process, and any error that occurred trying to close.
func (r *run_func_t) close() (error, error) {
	// If I haven't yet received a stop command, try and stop
	err := r.err.Get()
	if err == run_node_starting || err == run_node_running {
		if r.cmd != nil && r.cmd.Process != nil {
			_ = r.cmd.Process.Kill()
			// XXX What's the proper way to deal with an app
			// that's really not stopping?
		}
		r.cmd = nil
	}
	if r.wait != nil {
		r.wait.Wait()
		r.wait = nil
	}
	return r.err.Get(), nil
}

func streamOutput(r io.Reader, pinname string, output phly.NodeOutput) {
	buf := make([]byte, 256)
	for {
		size, err := r.Read(buf)
		if size > 0 {
			lines := getSteamOutput(string(buf[:size]))
			b := phly.PinBuilder{}.Add(pinname, phly.NewStringDoc(lines...))
			output.SendPins(b.Pins())
		}
		if err == io.EOF {
			return
		} else if err != nil {
			break
		}
	}
}

func getSteamOutput(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// ----------------------------------------
// CONST and VAR

const (
	run_cmdinput  = "cmd"
	run_clainput  = "cla"
	run_output    = "out"
	run_erroutput = "err"
)

var (
	run_node_starting = errors.New("rns")
	run_node_running  = errors.New("rnw")
)

func runFakeFmt() {
	fmt.Println()
}
