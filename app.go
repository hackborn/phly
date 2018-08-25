package phly

// Support for phly-based applications

import (
	"fmt"
	"github.com/micro-go/parse"
	"os"
	"os/signal"
	"sort"
	"syscall"
	//	"time"
)

func RunApp() (Pins, error) {
	filename, clas, err := readCla(os.Args)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return nil, nil
	}
	return runPipeline(filename, clas)
}

func runPipeline(filename string, clas map[string]string) (Pins, error) {
	p, err := LoadPipeline(filename)
	if err != nil {
		return nil, err
	}
	go func() {
		finished := make(chan os.Signal, 1)
		//		fmt.Println("signal")
		signal.Notify(finished, os.Interrupt, syscall.SIGTERM)
		//		fmt.Println("wait")
		<-finished
		//		fmt.Println("done 1")
		p.Stop()
		//		fmt.Println("done 2")
		//		time.Sleep(1000 * time.Millisecond)
		//		fmt.Println("done sleeping")
	}()

	// XXX Need to figure out how I get output back from the pin sender.
	args := StartArgs{Cla: clas}
	output := &pins{}
	err = p.Start(args)
	if err != nil {
		return nil, err
	}
	err = p.Wait()
	//	fmt.Println("sleep 1")
	//	time.Sleep(1000 * time.Millisecond)
	//	fmt.Println("app run done err", err)
	return output, err
}

func readCla(args []string) (string, map[string]string, error) {
	clas := make(map[string]string)
	token := parse.NewStringToken(args...)
	// Skip the app name
	token.Next()
	filename := ""
	for cur, err := token.Next(); err == nil; cur, err = token.Next() {
		// Handle commands
		switch cur {
		case "-vars":
			describeVars()
			return "", nil, nil
		case "-nodes":
			describeNodes()
			return "", nil, nil
		case "-markdown":
			markdownNodes()
			return "", nil, nil
		}
		// First token is the file
		if filename == "" {
			filename = cur
		} else {
			// All either args are CLA key / value pairs
			nxt, err := token.Next()
			if err != nil {
				return "", nil, err
			}
			clas[cur] = nxt
		}
	}
	// Default. Primarily for testing. Should probably make this configurable.
	if filename == "" {
		filename = `scaleimg.json`
		filename = `run.json`
	}
	return filename, clas, nil
}

func describeVars() {
	for _, v := range vardescrs {
		fmt.Println(v.name, "-", v.descr)
	}
}

func describeNodes() {
	for _, v := range sortedNodes() {
		descr := v.Describe()
		fmt.Println(descr.ClaString())
	}
}

func markdownNodes() {
	for _, v := range sortedNodes() {
		descr := v.Describe()
		fmt.Println(descr.MarkdownString())
	}
}

// --------------------------------
// SORT

func sortedNodes() []NodeFactory {
	var nodes []NodeFactory
	for _, v := range reg.factories {
		nodes = append(nodes, v)
	}
	sort.Sort(SortNodeFactory(nodes))
	return nodes
}

type SortNodeFactory []NodeFactory

func (s SortNodeFactory) Len() int {
	return len(s)
}
func (s SortNodeFactory) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortNodeFactory) Less(i, j int) bool {
	return s[i].Describe().Id < s[j].Describe().Id
}
