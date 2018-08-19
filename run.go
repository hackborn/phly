package phly

import (
	"bufio"
	"fmt"
	"github.com/micro-go/lock"
	"github.com/micro-go/parse"
	"io"
	"os/exec"
	"strings"
	"sync"
)

const (
	run_cmdinput = "cmd"
	run_clainput = "cla"
	run_output   = "out"
)

// run executes a command.
type run struct {
	cmd     *exec.Cmd
	wg      *sync.WaitGroup
	running lock.AtomicBool
	err     error
}

func (n *run) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/run", Name: "Run", Purpose: "Run a program."}
	descr.InputPins = append(descr.InputPins, PinDescr{Name: run_cmdinput, Purpose: "The command to run."})
	descr.InputPins = append(descr.InputPins, PinDescr{Name: run_clainput, Purpose: "Optional command line arguments."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: run_output, Purpose: "The file list."})
	return descr
}

func (n *run) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	running := lock.NewAtomicBool()
	return &run{running: running}, nil
}

func (n *run) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	err := n.Close()
	if err != nil {
		return nil, err
	}

	filename, cla, err := n.getInput(input)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return nil, BadRequestErr
	}
	// Support two formats -- either passing in an array of CLAs, or a single string that needs to be parsed.
	if len(cla) == 1 {
		cla = parse.AsArguments(cla[0])
	}
	n.cmd = exec.Command(args.Filename(filename), cla...)
	n.wg = &sync.WaitGroup{}
	n.wg.Add(1)
	go n.runCmd(n.cmd, n.wg, sender)
	return Running, nil
}

func (n *run) Close() error {
	fmt.Println("close 1")
	var err error
	if n.cmd != nil {
		fmt.Println("close 2")
		if n.running.IsTrue() {
			fmt.Println("kill?")
			err = n.cmd.Process.Kill()
			fmt.Println("kill err", err)
		}
		n.cmd = nil
	}
	if err != nil {
		// Not sure how to recover from this condition
		// If the app is really not stopping, I don't want to wait, but
		// if it's in the middle of stopping or going to stop, I need to.
		//		n.wg = nil
	}
	if n.wg != nil {
		n.wg.Wait()
		n.wg = nil
	}
	n.running.SetTo(false)
	n.err = nil
	return err
}

func (n *run) runCmd(cmd *exec.Cmd, wg *sync.WaitGroup, sender PinSender) {
	defer fmt.Println("runCmd done")
	defer wg.Done()
	defer n.running.SetTo(false)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sender.SendFinished(n, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		sender.SendFinished(n, err)
		return
	}
	go func() {
		streamOutput(stdout, n, sender)
	}()
	go func() {
		streamOutput(stderr, n, sender)
	}()

	err = cmd.Start()
	if err != nil {
		sender.SendFinished(n, err)
		return
	}

	n.running.SetTo(true)
	err = cmd.Wait()
	fmt.Println("finished wait", err)
	sender.SendFinished(n, err)
}

func streamOutput(r io.Reader, n Node, sender PinSender) {
	defer fmt.Println("finished stream")
	buf := make([]byte, 256)
	for {
		size, err := r.Read(buf)
		if size > 0 {
			fmt.Println("\toverheard:", string(buf[:size]))
			pins := NewPins()
			pins.Add(run_output, getOutput(string(buf[:size])))
			sender.SendPins(n, pins)
		}
		if err == io.EOF {
			return
		} else if err != nil {
			break
		}
	}
}

func (n *run) getInput(input Pins) (string, []string, error) {
	filenames := getItems(input.Get(run_cmdinput))
	if len(filenames) < 1 {
		return "", nil, BadRequestErr
	}
	args := getItems(input.Get(run_clainput))
	return filenames[0], args, nil
}

func getOutput(s string) *Doc {
	d := &Doc{}
	p := d.NewPage("")
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		p.AddItem(scanner.Text())
	}
	return d
}

func getItems(docs []*Doc) []string {
	if len(docs) > 0 {
		return docs[0].GetStringItems()
	}
	return nil
}
