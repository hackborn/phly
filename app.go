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
	filename, err := readCla(os.Args)
	if err != nil {
		return nil, err
	}
	if filename == "" {
		return nil, nil
	}
	return runPipeline(filename)
}

func runPipeline(filename string) (Pins, error) {
	p, err := LoadPipeline(filename)
	if err != nil {
		return nil, err
	}
	args := RunArgs{Env: env, Cla: os.Args, WorkingDir: workingDir(filename)}

	input := &pins{}
	output := &pins{}
	err = p.Run(args, input, output)
	return output, err
}

func readCla(args []string) (string, error) {
	token := parse.NewStringToken(args...)
	// Skip the app name
	token.Next()
	filename := ""
	for a, err := token.Next(); err == nil; a, err = token.Next() {
		switch a {
		case "-vars":
			describeVars()
			return "", nil
		case "-nodes":
			describeNodes()
			return "", nil
		case "-markdown":
			markdownNodes()
			return "", nil
		}
		// First token is the file
		if filename == "" {
			filename = a
		}
	}
	// Default. Primarily for testing. Should probably make this configurable.
	if filename == "" {
		filename = `scaleimg.json`
	}
	return filename, nil
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
