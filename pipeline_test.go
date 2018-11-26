package phly

import (
	"fmt"
	"strings"
	"testing"
)

// ----------------------------------------
// READ-PIPELINE

func TestReadPipeline(t *testing.T) {
	cases := []struct {
		Pipeline string
		WantErr  error
	}{
		{testPipelineBadData1, NewParseError(nil)},
		{testPipelineBadData2, NewMissingError("")},
		// XXX How should we do success tests?
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			p := &pipeline{}
			have_err := readPipeline(strings.NewReader(tc.Pipeline), p)
			if !ErrorsEqual(have_err, tc.WantErr) {
				fmt.Println("err mismatch\nhave\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			}
		})
	}
}

// ----------------------------------------
// RUN-PIPELINE

func TestRunPipeline(t *testing.T) {
	Register(&test_source_node{})

	cases := []struct {
		Pipeline string
		Input    Pins
		WantErr  error
	}{
		{testPipelineData1, nil, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			p := &pipeline{}
			have_err := readPipeline(strings.NewReader(tc.Pipeline), p)
			if have_err != nil {
				fmt.Println("err should be nil but is", have_err)
				t.Fatal()
			}
			clas := make(map[string]string)
			args := StartArgs{Cla: clas}
			have_err = p.Run(args, tc.Input)
			if !ErrorsEqual(have_err, tc.WantErr) {
				fmt.Println("error mismatch\nhave\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			}
			//			t.Fatal()
		})
	}
}

// ----------------------------------------
// TEST-SOURCE-NODE

const (
	testnode_start = "start"
	testnode_in    = "in"
	testnode_out   = "out"
)

// test_source_node is used solely in tests. It generates a value in the start and finishes.
type test_source_node struct {
	Items []interface{} `json:"items,omitempty"`
}

func (n *test_source_node) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/test/source", Name: "Test Source", Purpose: "A source node for running tests."}
	//	descr.InputPins = append(descr.InputPins, PinDescr{Name: testnode_in, Purpose: "Input."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: testnode_out, Purpose: "Output."})
	return descr
}

func (n *test_source_node) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &test_source_node{}, nil
}

func (n *test_source_node) Process(args ProcessArgs, stage NodeStage, input Pins, output NodeOutput) error {
	fmt.Println("TEST NODE PROCESS", stage, "send", n.Items)
	if len(n.Items) > 0 {
		cmds := []interface{}{testnode_out}
		cmds = append(cmds, n.Items...)
		pins, err := BuildPins(cmds...)
		if err != nil {
			return err
		}
		output.SendPins(pins)
	}
	output.SendMsg(MsgFromStop(nil))
	return nil
}

func (n *test_source_node) StopNode(args StoppedArgs) error {
	fmt.Println("testing node stopped!")
	return nil
}

// ----------------------------------------
// CONST and VAR

const (
	// Search for 'Z' to see where the bad data is broken
	testPipelineBadData1 = `{ "nodes": Z{ "test1": { "node": "phly/test/source", "cfg": { "runmode": "" } } } }`
	testPipelineBadData2 = `{ "nodes": { "test1": { "node": "Zphly/test/source", "cfg": { "runmode": "" } } } }`

	testPipelineData1 = `{
	"nodes": {
		"test1": {
			"node": "phly/test/source",
			"cfg": {
				"runmode": "running",
				"finish": ["finish"]
			}
		}
	}
}`
)
