package phly

// Support for phly-based applications

import (
	"fmt"
	"github.com/micro-go/parse"
	"os"
	"sort"
)

func RunApp() (PipelineResult, error) {
	filename := readCla(os.Args)
	if filename == "" {
		return PipelineResult{}, nil
	}
	return runPipeline(filename)
}

func runPipeline(filename string) (PipelineResult, error) {
	p, err := LoadPipeline(filename)
	if err != nil {
		return PipelineResult{}, err
	}
	args := RunArgs{}
	return p.Run(args)
}

func readCla(args []string) string {
	filename := `.\data\scale_image.json`
	if args == nil {
		return filename
	}
	token := parse.NewStringToken(args...)
	for a, err := token.Next(); err == nil; a, err = token.Next() {
		switch a {
		case "nodes":
			describeNodes()
			filename = ""
		case "markdown":
			markdownNodes()
			filename = ""
		}
	}
	return filename
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
