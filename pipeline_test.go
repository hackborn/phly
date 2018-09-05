package phly

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestRunPipeline1(t *testing.T) {
	// XXX Not working
	//	testRunAndWaitPipeline(t, testPipelineData1)
}

func TestRunPipeline2(t *testing.T) {
	testRunAndWaitPipeline(t, testPipelineData2)
}

func TestRunPipeline3(t *testing.T) {
	testFreeRunPipeline(t, testPipelineData3, time.Millisecond*10)
}

func testRunAndWaitPipeline(t *testing.T, src string) {
	Register(&testnode{})

	p, err := ReadPipeline(strings.NewReader(src))
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	clas := make(map[string]string)
	args := StartArgs{Cla: clas}
	err = p.Start(args)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
	err = p.Wait()
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
}

func testFreeRunPipeline(t *testing.T, src string, dur time.Duration) {
	Register(&testnode{})

	p, err := testReadPipeline(strings.NewReader(src))
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	go func() {
		time.Sleep(dur)
		p.Stop()
	}()

	clas := make(map[string]string)
	args := StartArgs{Cla: clas}
	err = p.Start(args)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
	err = p.Wait()
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
}

func testReadPipeline(r io.Reader) (*pipeline, error) {
	p := &pipeline{}
	err := readPipeline(r, p)
	return p, err
}

const (
	testPipelineData1 = `{
	"nodes": {
		"test1": {
			"node": "phly/test",
			"cfg": {
				"runmode": ""
			}
		}
	}
}`

	testPipelineData2 = `{
	"nodes": {
		"test1": {
			"node": "phly/test",
			"cfg": {
				"runmode": "",
				"finish": ["finish"]
			}
		}
	}
}`

	testPipelineData3 = `{
	"nodes": {
		"test1": {
			"node": "phly/test",
			"cfg": {
				"runmode": "running",
				"finish": ["finish"]
			}
		}
	}
}`
)
