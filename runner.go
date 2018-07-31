package phly

import (
	"fmt"
	"reflect"
)

type runner struct {
}

// run is the basic node-running algorithm. It loops over the
// current nodes, running each one, then distributing the outputs.
func (r *runner) run(args RunArgs, pipe *pipeline, nodes []*container) (Pins, error) {
	if len(nodes) < 1 {
		return nil, BadRequestErr
	}

	stack := newRunnerStack(pipe, nodes)
	current, err := stack.popNext()
	if err != nil {
		return nil, err
	}
	for current != nil {
		// Run the node
		output := &pins{}
		err := current.run(args, output)
		if err != nil {
			return nil, err
		}

		// Distribute output
		err = current.distributeOutput(output, stack)
		if err != nil {
			return nil, err
		}

		// Get next
		current, err = stack.popNext()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// --------------------------------
// RUNNER-CONTAINER

// runnerContainer wraps a pipeline container with behaviour
// for tracking and storing the inputs.
type runnerContainer struct {
	c      *container
	inputs []runnerConnection
}

func (c *runnerContainer) isReady() bool {
	return len(c.inputs) == len(c.c.inputs)
}

func (c *runnerContainer) run(args RunArgs, output Pins) error {
	fmt.Println("run", reflect.TypeOf(c.c.node))
	return c.c.node.Run(args, c.makeInput(), output)
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

func (c *runnerContainer) distributeOutput(outs *pins, stack *runnerStack) error {
	for name, docs := range outs.all {
		// Get all destinations for this pin. If we just have one, we
		// can pass along the data, otherwise it should be copied.
		dests := c.getDestinations(name, stack)
		for i, dst := range dests {
			newDocs := docs
			if i > 0 {
				newDocs = copyDocs(docs)
			}
			dst.rc.addInput(dst.pin, newDocs)
		}
	}
	return nil
}

func (c *runnerContainer) getDestinations(name string, stack *runnerStack) []runnerDist {
	// Compile all node connections.
	var ans []runnerDist
	for _, out := range c.c.outputs {
		if name == out.srcPin {
			ans = append(ans, runnerDist{out.dstPin, stack.add(out.dstNode)})
		}
	}
	return ans
}

func (c *runnerContainer) addInput(pin string, docs []*Doc) {
	c.inputs = append(c.inputs, runnerConnection{pin, docs})
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
// RUNNER-STACK

type runnerStack struct {
	pipe  *pipeline
	nodes []*runnerContainer
}

func newRunnerStack(pipe *pipeline, nodes []*container) *runnerStack {
	r := &runnerStack{pipe: pipe}
	for _, n := range nodes {
		r.add(n)
	}
	return r
}

// add() adds the container to the stack, preventing duplicates.
// Answer the runnerContainer.
func (r *runnerStack) add(c *container) *runnerContainer {
	for _, found := range r.nodes {
		if c == found.c {
			return found
		}
	}
	ans := &runnerContainer{c: c}
	r.nodes = append(r.nodes, ans)
	return ans
}

// popNext() answer the next node that's ready to run.
func (r *runnerStack) popNext() (*runnerContainer, error) {
	if len(r.nodes) < 1 {
		return nil, nil
	}
	for i, n := range r.nodes {
		if n.isReady() {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			return n, nil
		}
	}
	return nil, unfinishedPipelineErr
}

func fakeFmt() {
	fmt.Println("")
}
