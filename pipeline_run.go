package phly

import (
	"errors"
	"fmt"
	"github.com/micro-go/lock"
	"reflect"
	"sync"
	"unsafe"
)

var (
	pid_counter = lock.NewAtomicInt32()
)

// ----------------------------------------
// PIPELINE-RUNNER

// pipeline_runner struct is the state of a currently running pipeline.
type pipeline_runner struct {
	sargs       StartArgs
	pargs       ProcessArgs
	done        chan struct{}
	p           *pipeline
	wait        *sync.WaitGroup
	msgchan     chan *pipeline_msg
	passthrough chan *pipeline_msg // Message channel that acts as a buffer, preventing cases where a node would send a message on the main thread and block.
	err         lock.AtomicError   // Store the current state of the running operation, or its result.
	pid         int32
}

func startPipelineRunner(p *pipeline, sargs StartArgs, pargs ProcessArgs, input Pins) (*pipeline_runner, error) {
	done := make(chan struct{})
	msgchan := make(chan *pipeline_msg, 128)
	passthrough := make(chan *pipeline_msg, 128)
	runner := &pipeline_runner{sargs: sargs, pargs: pargs, done: done, p: p, wait: &sync.WaitGroup{}, msgchan: msgchan, passthrough: passthrough, err: lock.NewAtomicError()}
	runner.pid = pid_counter.Add(1)
	starting, err := runner.getInitialInputs(input)
	if err != nil {
		return nil, err
	}
	if starting.empty() {
		return nil, NewIllegalError("No initial nodes")
	}

	runner.err.SetTo(pipeline_starting)
	runner.wait.Add(1)
	go runner.runPassthrough()
	go runner.run(pargs, starting)
	return runner, nil
}

// close() closes the runner, answering the error returned.
func (r *pipeline_runner) close() error {
	if r.done != nil {
		close(r.done)
		r.done = nil
	}
	if r.wait != nil {
		r.wait.Wait()
		r.wait = nil
	}
	// Close message channels after we've waited for all routines to stop,
	if r.passthrough != nil {
		close(r.passthrough)
		r.passthrough = nil
	}
	if r.msgchan != nil {
		close(r.msgchan)
		r.msgchan = nil
	}
	return r.err.Get()
}

func (r *pipeline_runner) runPassthrough() {
	for {
		select {
		case <-r.done:
			return
		case msg, more := <-r.passthrough:
			if more {
				r.msgchan <- msg
			}
		}
	}
}

func (r *pipeline_runner) run(args ProcessArgs, starting *nodeInputs) {
	defer fmt.Println("***RUN DONE")
	fmt.Println("***+++++++++++++RUN STARTED pid", r.pid)
	var err error

	defer r.runFinished()
	defer r.wait.Done()
	defer func() { r.err.SetTo(err) }()

	state := newPipelineRunningState(r.p, args, r.passthrough)
	defer state.stopAll()
	/*
		err = state.start(starting)
		fmt.Println("start err", err)
		if err != nil {
			return
		}
	*/

	fmt.Println("p.run() for")
	// Treat the initial inputs like any input comimg into
	// the system and queue them up
	for name, ins := range starting.nodes {
		r.passthrough <- newPipelineMsg(Msg{What: WhatPins, Payload: ins}, name)
	}

	r.err.SetTo(pipeline_running)
	for {
		select {
		case <-r.done:
			return
		case msg, more := <-r.msgchan:
			if more {
				//				fmt.Println("MSG!", msg, "for", int64(r))
				fmt.Println("MSG!", msg, "for", r.pid)
				err, more = r.runMsg(state, msg)
				if err != nil || !more {
					fmt.Println("end run from err", err, "more", more)
					return
				}
			}
		}
	}
}

// runFinished() notifies the parent that the runner is ending.
func (r *pipeline_runner) runFinished() {
	fmt.Println("run finished")
	if r.sargs.output != nil {
		output := r.sargs.output.(*pipelineNodeOutput)
		fmt.Println("SEND TO OUTPUT", *((*int32)(unsafe.Pointer(output))))
		r.sargs.output.SendMsg(MsgFromStop(nil))
		r.sargs.output.SendMsg(Msg{What: whatFlush, Payload: nil})
	}
}

func (r *pipeline_runner) runMsg(state *pipeline_running_state, msg *pipeline_msg) (error, bool) {
	// Deal with any nodes that set the stop after it could be handled but
	// before the pins were received.
	state.removeStopped()

	var err error
	switch msg.What {
	case WhatPins:
		// Nil pins are valid, in the case of source nodes running at startup
		var p Pins
		if msg.Payload != nil {
			_p, ok := msg.Payload.(Pins)
			if !ok {
				// XXX Warning? Error?
				return nil, true
			}
			p = _p
		}
		err = r.runPins(state, msg.Node, p)
		/*
			case whatFlush:
				// A flush happens when I have nothing left to process -- I send a message
				// down the channel to make sure no one is in flight that will start new nodes.
				//		fmt.Println("HANDLE FLUSH")
				if state.empty() {
					fmt.Println("flush state is empty")
					return nil, false
				}
		*/
	}

	// If no messages are in flight but no one is processing, I'm stuck, and need to quit.
	if msg.id == msg_id.Get() && state.waiting() {
		fmt.Println("ON FINAL MSG", msg_id.Get(), "AND DONE")
		// XXX This is probably an error? I need to get a sense of the success conditions to know.
		return nil, false
	}
	return err, true
}

func (r *pipeline_runner) runPins(state *pipeline_running_state, nodename string, _pins Pins) error {
	fmt.Println("HANDLE PINS!", nodename, _pins)

	var err error

	// Look up the node, creating if necessary, then feeding in the pins.
	err = state.process(nodename, _pins)

	// Check if we need to stop
	state.removeStopped()
	if state.empty() {
		fmt.Println("send flush")
		r.passthrough <- newPipelineMsg(Msg{What: whatFlush, Payload: nil}, "")
	}
	return err
}

// ----------------------------------------
// PIPELINE-RUNNER INITIALIZE

// getInitialInputs() gets all nodes and corresponding inputs that should
// be performed at the initial processing. Note that a node doesn't need
// input to be included: It might be a source node with no inputs.
func (p *pipeline_runner) getInitialInputs(input Pins) (*nodeInputs, error) {
	ins := nodeInputs{}
	fmt.Println("Get initial on SIZE", len(p.p.nodes))
	// 1. All nodes with no input
	err := p.getSourceInputs(&ins)

	// 2. All nodes with input connected to the pipeline
	//	sn := p.getNodesForInput(input)

	// 3. All nodes with input from the args or containing pipeline.
	err = MergeErrors(err, p.getSpecialnputs(input, &ins))

	return &ins, err
}

// getSourceNodes() answers all nodes with no input.
func (p *pipeline_runner) getSourceInputs(ins *nodeInputs) error {
	if len(p.p.nodes) < 1 {
		return nil
	}
	for n, c := range p.p.nodes {
		if len(c.inputs) < 1 {
			ins.add(n, "", nil)
		}
	}
	return nil
}

// getSpecialnputs() answers all nodes with argument or pipeline input.
// input is the input received by my calling pipeline.
func (p *pipeline_runner) getSpecialnputs(input Pins, ins *nodeInputs) error {
	if len(p.p.nodes) < 1 {
		return nil
	}

	for _, dstn := range p.p.nodes {
		for _, conn := range dstn.inputs {
			if conn.dstNode.name == args_container.name {
				doc := p.p.args.valueDoc(p.pargs, conn.dstPin)
				if doc != nil {
					ins.add(dstn.name, conn.srcPin, NewDocs(doc))
				}
			} else if conn.dstNode.name == pipeline_container.name && input != nil {
				// Mape the pipeline input to this node's input.
				docs := input.GetPin(conn.dstPin)
				if len(docs.Docs) > 0 {
					ins.add(dstn.name, conn.srcPin, &docs)
				}
			}
		}
	}
	return nil
}

// getNodesForInput() answers all nodes targeted by the pins.
func (p *pipeline_runner) getNodesForInput(input Pins, starting *nodeInputs) {
	// For each input in the pins, find the destination node and add it to the incoming stack.
	if input == nil {
		return
	}
	//		for _, descr := range p.inputDescr {
	//			docs := src.Remove(descr.Name)
	//			if docs != nil {
	//				for _, conn := range descr.connections {
	//					ri.add(conn.DstNode, conn.DstPin, docs)
	//				}
	//			}
	//		}
}

// ----------------------------------------
// PIPELINE-RUNNING-STATE

// pipeline_running_state struct stores the state of the running performer thread.
type pipeline_running_state struct {
	p       *pipeline
	args    ProcessArgs
	nodes   map[string]*pipeline_running_node
	msgchan chan *pipeline_msg
}

func newPipelineRunningState(p *pipeline, args ProcessArgs, msgchan chan *pipeline_msg) *pipeline_running_state {
	nodes := make(map[string]*pipeline_running_node)
	return &pipeline_running_state{p, args, nodes, msgchan}
}

func (p *pipeline_running_state) empty() bool {
	return len(p.nodes) < 1
}

// waiting() returns true if I can't currently process anything.
func (p *pipeline_running_state) waiting() bool {
	fmt.Println("waiting len", len(p.nodes))
	for _, v := range p.nodes {
		fmt.Println("\tstage is", v.stage)
		if v.stage != NodeStarting {
			return false
		}
	}
	return true
}

func (p *pipeline_running_state) stopAll() {
	for k, v := range p.nodes {
		v.node.StopNode(StoppedArgs{})
		delete(p.nodes, k)
	}
}

func (p *pipeline_running_state) removeStopped() {
	for k, v := range p.nodes {
		if v.output.stopped.IsTrue() {
			v.node.StopNode(StoppedArgs{})
			delete(p.nodes, k)
		}
	}
}

// handlePins() starts each node in the collection (if not currently running) and
// supplies the pins.
// nodename is the node that should process the pins.
func (p *pipeline_running_state) process(nodename string, pins Pins) error {
	fmt.Println("handle pins for", nodename)
	// Find node
	n := p.nodes[nodename]

	// Create if missing
	if n == nil {
		container := p.p.nodes[nodename]
		if container == nil {
			return NewMissingError("Node " + nodename)
		}
		// XXX We're passing in the pipeline for the output resolver, but probably
		// we should be caching all the state from the pipeline that we use
		n = newPipelineRunningNode(p.args, container, p.msgchan, p.p)
		p.nodes[nodename] = n
	}

	// Feed input
	return n.process(pins)
}

// ----------------------------------------
// PIPELINE-RUNNING-NODE

// pipeline_running_node struct stores an actively running node.
type pipeline_running_node struct {
	args     ProcessArgs
	node     Node
	output   *pipelineNodeOutput
	stage    NodeStage
	starting *node_starting // Determine when a node moves from starting to running
}

func newPipelineRunningNode(args ProcessArgs, container *container, msgchan chan *pipeline_msg, resolver outputResolver) *pipeline_running_node {
	fmt.Println("run", container.name, reflect.TypeOf(container.node))
	output := newPipelineNodeOutput(container.name, msgchan, resolver)
	n := &pipeline_running_node{args, container.node, output, NodeStarting, &node_starting{}}
	return n
}

func (n *pipeline_running_node) process(pins Pins) error {
	// We are either accumulating all input needed to start, or we are running.
	if n.stage == NodeStarting {
		n.starting.accumulate(pins)
		if n.starting.ready() {
			n.stage = NodeRunning
			return n.node.Process(n.args, NodeStarting, &n.starting.pins, n.output)
		}
	} else if pins != nil {
		return n.node.Process(n.args, n.stage, pins, n.output)
	}
	return nil
	// So this is pretty complicated. It needs to do a lot of things:

	// While stage == waiting, accumulate starting input, if any.
	// When all starting input is accumulated, send it in with starting.
	// If there are no inputs, and stage = waiting, start the processing.
	// If stage == running, send in any input. But here, there must be
	// input. Only if stage is waiting and the node has no inputs can Pins be nil.

	// So... how to capture all that?

	// A lot of it comes down to a "waiting" state, I guess. As soon as we've
	// moved to running, we're a simple passthrough. And then a lot of the
	// complexity comes down to this accumulation behaviour.

	// So... an object that does the accumulation, reporting when it's safe to
	// call starting based on the state of the pins? Sounds fair

}

// ----------------------------------------
// NODE-INPUTS

// nodeInputs stores a map of nodes to their input pin data.
// It is structurally the same as nodeOutputs, but semantically
// (and behaviourly) different.
// XXX Actually it looks like it's identical...
type nodeInputs struct {
	nodes map[string]*pins
}

func (s nodeInputs) empty() bool {
	if s.nodes == nil || len(s.nodes) < 1 {
		return true
	}
	return false
}

func (s *nodeInputs) add(node, pin string, docs *Docs) {
	if s.nodes == nil {
		s.nodes = make(map[string]*pins)
	}
	in, ok := s.nodes[node]
	if !ok {
		in = &pins{}
		s.nodes[node] = in
	}
	// XXX We should be smart about copying when necessary
	in.addDocs(pin, docs)
}

// ----------------------------------------
// NODE-OUTPUTS

// nodeOutputs stores a map of nodes to their output pin data.
// It is structurally the same as nodeInputs, but semantically
// (and behaviourly) different.
// XXX Actually it looks like it's identical...
type nodeOutputs struct {
	nodes map[string]*pins
}

func (s nodeOutputs) empty() bool {
	if s.nodes == nil || len(s.nodes) < 1 {
		return true
	}
	return false
}

func (s *nodeOutputs) add(node, pin string, docs *Docs) {
	if s.nodes == nil {
		s.nodes = make(map[string]*pins)
	}
	in, ok := s.nodes[node]
	if !ok {
		in = &pins{}
		s.nodes[node] = in
	}
	// XXX We should be smart about copying when necessary
	in.addDocs(pin, docs)
}

// ----------------------------------------
// PIPELINE-NODE-OUTPUT

// pipelineNodeOutput is used to deal with node output happening
// inside of the pipeline's graph.
type pipelineNodeOutput struct {
	name     string
	stopped  lock.AtomicBool
	msgchan  chan<- *pipeline_msg
	resolver outputResolver
}

func newPipelineNodeOutput(name string, msgchan chan<- *pipeline_msg, resolver outputResolver) *pipelineNodeOutput {
	return &pipelineNodeOutput{name, lock.NewAtomicBool(), msgchan, resolver}
}

func (p *pipelineNodeOutput) SendPins(pins Pins) {
	p.SendMsg(MsgFromPins(pins))
}

func (p *pipelineNodeOutput) SendMsg(msg Msg) {
	if msg.What == WhatPins {
		p.handlePinOutputs(msg)
	} else if msg.What == WhatStop {
		// Stop is handled differently -- the framework immediately pulls
		// the node from the processing graph. This allows the possibility for
		// one-shot nodes to be immediately restarted in a loopback.
		p.stopped.SetTo(true)
	} else {
		p.msgchan <- newPipelineMsg(msg, p.name)
	}
}

// handlePinOutputs() converts the outputs into node inputs and feeds them into
// the system.
func (p *pipelineNodeOutput) handlePinOutputs(msg Msg) {
	fmt.Println("handlePinOutputs for", p.name)
	if msg.What != WhatPins {
		return
	}
	pins, ok := msg.Payload.(Pins)
	if !ok || pins == nil {
		return
	}
	fmt.Println("handlePinOutputs - walk")
	pins.WalkPins(func(name string, docs Docs) {
		// I need destination node and pin names
		dstnode, dstpin, err := p.resolver.ResolveOutput(p.name, name)
		fmt.Println("\thandlePinOutputs - dst", dstnode, dstpin, err)
		outpins, err := BuildPins(dstpin, &docs)
		if outpins == nil || err != nil {
			return
		}
		fmt.Println("\thandlePinOutputs 2 - dst", dstnode, dstpin, err)
		p.msgchan <- newPipelineMsg(MsgFromPins(outpins), dstnode)
	})
}

// ----------------------------------------
// NODE-STARTING

// node_starting contains state and behaviour to determine when a newly-added
// node can receive the NodeStarting stage and input pins.
type node_starting struct {
	pins pins
}

func (n *node_starting) accumulate(pins Pins) {
	if pins != nil {
		pins.WalkPins(func(name string, docs Docs) {
			n.pins.addDocs(name, &docs)
		})
	}
}

// ready() answers true if my data matches the conditions required
// by the underlying node to start.
func (n *node_starting) ready() bool {
	return true
}

// ----------------------------------------
// PIPELINE-MSG

type pipeline_msg struct {
	id int32
	Msg
	Node string
}

func newPipelineMsg(m Msg, n string) *pipeline_msg {
	return &pipeline_msg{msg_id.Add(1), m, n}
}

// ----------------------------------------
// OUTPUT-RESOLVER

// outputResolver is used to convert a source node and output pin name
// to a destination node and input pin name
type outputResolver interface {
	ResolveOutput(srcnode, srcpin string) (string, string, error)
}

// ----------------------------------------
// CONST and VAR

var (
	pipeline_starting = errors.New("ps")
	pipeline_running  = errors.New("pr")

	msg_id = lock.NewAtomicInt32()
)

func pipelineRunFakeFmt() {
	fmt.Println()
}
