package main

import (
	"fmt"
	"github.com/hackborn/phly"
	_ "github.com/hackborn/phly_img"
)

func main() {
	doc := &phly.Doc{}
	doc.SetInt("boop2", 45)
	doc.SetString("boop", "bop")

	_, err := phly.RunPipeline()
	if err != nil {
		fmt.Println(err)
	}
}
