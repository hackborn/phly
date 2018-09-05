package phly

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunPipeline1(t *testing.T) {
	// XXX Not working
	// testRunAndWaitPipeline(t, testPipelineData1)
}

func TestRunPipeline2(t *testing.T) {
	testRunAndWaitPipeline(t, testPipelineData2)
}

func TestRunPipeline3(t *testing.T) {
	testFreeRunPipeline(t, testPipelineData3, time.Millisecond*10)
}

func testRunAndWaitPipeline(t *testing.T, src string) {
	Register(&testnode{})

	p, err := testReadPipeline(strings.NewReader(src))
	testCheckFailed(t, err)

	clas := make(map[string]string)
	args := StartArgs{Cla: clas}

	err = p.Start(args)
	testCheckFailed(t, err)

	err = p.Wait()
	testCheckFailed(t, err)
	testCheckFailed(t, testPipelineStoppedErr(p, runFinishedStopped))
}

func testFreeRunPipeline(t *testing.T, src string, dur time.Duration) {
	Register(&testnode{})

	p, err := testReadPipeline(strings.NewReader(src))
	testCheckFailed(t, err)

	go func() {
		time.Sleep(dur)
		p.Stop()
	}()

	clas := make(map[string]string)
	args := StartArgs{Cla: clas}
	err = p.Start(args)
	testCheckFailed(t, err)

	err = p.Wait()
	testCheckFailed(t, err)
	testCheckFailed(t, testPipelineStoppedErr(p, requestedStopped))
}

func testReadPipeline(r io.Reader) (*pipeline, error) {
	p := &pipeline{}
	err := readPipeline(r, p)
	return p, err
}

func testCheckFailed(t *testing.T, err error) {
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func testPipelineStoppedErr(p *pipeline, required int32) error {
	if p.stopped.Get() != required {
		return errors.New("Stopped must be " + strconv.Itoa(int(required)))
	}
	return nil
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
