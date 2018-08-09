package phly

// Support for phly-based applications

import (
	"fmt"
	"github.com/micro-go/parse"
	"os"
	"path/filepath"
	"sort"
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
	args := RunArgs{Env: env, WorkingDir: workingDir(filename), cla: clas}

	input := &pins{}
	output := &pins{}
	err = p.Run(args, input, output)
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

func workingDir(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return filepath.Dir(path)
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
