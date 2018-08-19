package phly

import (
	"fmt"
	"github.com/micro-go/lock"
	"reflect"
	//	"sync"
)

type runner struct {
}

type runnerAsync struct {
	//	wg      sync.WaitGroup
	stopped lock.AtomicBool
	// These are not lock protected -- the design only has them read or written from one thread at a time.
	outs Pins
	err  error
	// Worker thread
	pipe   *pipeline
	active map[interface{}]*runnerContainer
}

// run is the basic node-running algorithm. It loops over the
// current nodes, running each one, then distributing the outputs.
func (r *runner) runAsync(args RunArgs, pipe *pipeline, nodes []*container, input runnerInput) (Pins, error) {
	if args.stop == nil {
		return nil, BadRequestErr
	}
	if len(nodes) < 1 {
		return nil, BadRequestErr
	}

	ro := newRunnerAsync(pipe, nodes, input)
	pipe.wg.Add(1)
	go ro.run(args, pipe, nodes, input)
	return ro.outs, ro.err
	//	return r.runAsyncWait(args, pipe, nodes, input)
}

/*
func (r *runner) runAsyncWait(args RunArgs, pipe *pipeline, nodes []*container, input runnerInput) (Pins, error) {
	ro := newRunnerAsync(pipe, nodes, input)
	ro.wg.Add(1)
	go ro.run(args, pipe, nodes, input)
	ro.wg.Wait()
	return ro.outs, ro.err
}
*/

func newRunnerAsync(pipe *pipeline, nodes []*container, input runnerInput) *runnerAsync {
	stopped := lock.NewAtomicBool()
	active := make(map[interface{}]*runnerContainer)

	r := &runnerAsync{stopped: stopped, pipe: pipe, active: active}
	for _, n := range nodes {
		rc := r.activate(n)
		if len(input.nodeInputs) > 0 {
			for k, v := range input.nodeInputs {
				if k == rc.c.name && v != nil {
					for pin, docs := range v.all {
						rc.addInput(pin, docs)
					}
				}
			}
		}
	}
	return r
}

// activate() adds the container to the stack, preventing duplicates.
// Answer the runnerContainer.
func (r *runnerAsync) activate(c *container) *runnerContainer {
	if found, ok := r.active[c.node]; ok && found != nil {
		return found
	}
	ans := newRunnerContainer(c)
	r.active[c.node] = ans
	return ans
}

func (r *runnerAsync) run(args RunArgs, pipe *pipeline, nodes []*container, input runnerInput) {
	defer fmt.Println("runner.run done 1")
	defer pipe.wg.Done()
	defer fmt.Println("runner.run done 2")
	defer r.closeActive()
	defer fmt.Println("runner.run done 3")

	sender := newRunnerPinSender()

	// Capture if my caller wants to end the main graph processing
	go func() {
		for {
			select {
			case <-args.stop:
				fmt.Println("set stoppd to true")
				r.stopped.SetTo(true)
				return
			}
		}
	}()

	// Do an initial pass of processing
	r.runActive(args, sender)
	if r.err != nil {
		return
	}

	// Process incoming messages until I have no running nodes.
	for {
		select {
		case <-args.stop:
			fmt.Println("stop")
			return
		case msg, _ := <-sender.c:
			if msg != nil && msg.node != nil {
				if msg.pins != nil {
					r.receiveOutput(msg.node, msg.pins)
					r.runActive(args, sender)
				}
				if msg.flow != nil {
					if msg.flow == Finished {
						r.closeNode(msg.node)
					}
				}
				if msg.err != nil {
					r.err = MergeErrors(r.err, msg.err)
				}
				if r.err != nil {
					return
				}
				if len(r.active) < 1 {
					return
				}
			}
		}
	}
}

// runActive() runs all active nodes.
func (r *runnerAsync) runActive(args RunArgs, sender PinSender) {
	for key, rc := range r.active {
		if r.stopped.IsTrue() {
			return
		}
		if !rc.dirty {
			continue
		}

		// Run the node
		flow, err := rc.run(args, sender)
		if err != nil {
			r.err = MergeErrors(r.err, err)
			return
		}
		rc.dirty = false
		if flow == Finished {
			fmt.Println("delete active")
			delete(r.active, key)
		}
	}
}

func (r *runnerAsync) closeNode(key interface{}) {
	if rc, ok := r.active[key]; ok && rc != nil {
		rc.c.close()
		fmt.Println("closeNode delete active")
		delete(r.active, key)
	}
}

func (r *runnerAsync) closeActive() {
	fmt.Println("close active 1")
	for key, _ := range r.active {
		fmt.Println("close node 1")
		r.closeNode(key)
		fmt.Println("close node 2")
	}
	fmt.Println("close active 2")
}

func (r *runnerAsync) receiveOutput(node interface{}, pins Pins) {
	if node == nil || pins == nil {
		return
	}
	//	fmt.Println("Someone sent me something!")
	c := r.findNodeContainer(node)
	if c == nil {
		// A bit of an error, here, but not sure how to handle it
		return
	}

	// Distribute outputs
	pins.Walk(func(name string, docs []*Doc) {
		// Get all destinations for this pin. If we just have one, we
		// can pass along the data, otherwise it should be copied.
		dests := getRunnerDestinations(name, c, r)
		for i, dst := range dests {
			newDocs := docs
			if i > 0 {
				newDocs = copyDocs(docs)
			}
			dst.rc.addInput(dst.pin, newDocs)
		}
	})
}

// findNodeContainer() answers the node container, either from my running list,
// or my source pipeline.
func (r *runnerAsync) findNodeContainer(node interface{}) *container {
	if rc, ok := r.active[node]; ok && rc != nil && rc.c != nil {
		return rc.c
	}
	for _, c := range r.pipe.nodes {
		if c.node == node {
			return c
		}
	}
	return nil
}

// --------------------------------
// RUNNER-INPUT

// runnerInput accumulates input data to a node.
type runnerInput struct {
	nodeInputs map[string]*pins
}

func (r *runnerInput) add(dstnode, dstpin string, docs []*Doc) {
	if docs == nil {
		return
	}
	if r.nodeInputs == nil {
		r.nodeInputs = make(map[string]*pins)
	}
	in, ok := r.nodeInputs[dstnode]
	if !ok {
		in = &pins{}
		r.nodeInputs[dstnode] = in
	}
	for _, doc := range docs {
		// XXX We should be smart about copying when necessary
		in.Add(dstpin, doc)
	}
}

func (r *runnerInput) hasPin(node, pin string) bool {
	if r.nodeInputs == nil {
		return false
	}
	if pins, ok := r.nodeInputs[node]; ok && pins != nil {
		return pins.Get(pin) != nil
	}
	return false
}

// --------------------------------
// RUNNER-CONTAINER

// runnerContainer wraps a pipeline container with behaviour
// for tracking and storing the inputs.
type runnerContainer struct {
	dirty  bool
	c      *container
	inputs []runnerConnection
}

func newRunnerContainer(c *container) *runnerContainer {
	return &runnerContainer{dirty: true, c: c}
}

func (c *runnerContainer) run(args RunArgs, sender PinSender) (Flow, error) {
	fmt.Println("run", reflect.TypeOf(c.c.node))
	return c.c.node.Run(args, c.makeInput(), sender)
}

func (c *runnerContainer) makeInput() Pins {
	p := &pins{}
	for _, in := range c.inputs {
		for _, doc := range in.docs {
			p.Add(in.pin, doc)
		}
	}
	return p
}

func (c *runnerContainer) addInput(pin string, docs []*Doc) {
	c.inputs = append(c.inputs, runnerConnection{pin, docs})
	c.dirty = true
}

func getRunnerDestinations(name string, c *container, r *runnerAsync) []runnerDist {
	// Compile all node connections.
	var ans []runnerDist
	for _, out := range c.outputs {
		if name == out.srcPin {
			ans = append(ans, runnerDist{out.dstPin, r.activate(out.dstNode)})
		}
	}
	return ans
}

// --------------------------------
// RUNNER-CONNECTION

// runnerConnection is an active input, with data.
type runnerConnection struct {
	pin  string
	docs []*Doc
}

// --------------------------------
// RUNNER-DIST

// runnerDist is a small class to help shuttle output data to input pins.
type runnerDist struct {
	pin string
	rc  *runnerContainer
}

// --------------------------------
// RUNNER-NODE-OUTPUT

type runnerPinSender struct {
	c chan *nodePinMsg
}

func newRunnerPinSender() *runnerPinSender {
	c := make(chan *nodePinMsg, 128)
	return &runnerPinSender{c}
}

func (r *runnerPinSender) SendPins(n Node, p Pins) {
	r.c <- &nodePinMsg{n, nil, p, nil}
}

func (r *runnerPinSender) SendFinished(n Node, err error) {
	r.c <- &nodePinMsg{n, Finished, nil, err}
}

// --------------------------------
// NODE-OUTPUT-MSG

type nodePinMsg struct {
	node interface{}
	flow Flow
	pins Pins
	err  error
}

func fakeFmt() {
	fmt.Println("")
}
