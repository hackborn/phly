package main

import (
	"fmt"
	"github.com/hackborn/phly"
	_ "github.com/hackborn/phly_img"
)

func main() {
	doc := &phly.Doc{}
	doc.Header.SetInt("boop2", 45)
	doc.Header.SetString("boop", "bop")
	i, ok := doc.Header.GetInt("boop2")
	fmt.Println("int", i, ok)
	str, ok := doc.Header.GetString("boop")
	fmt.Println("string", str, ok)

	_, err := phly.RunPipeline()
	if err != nil {
		fmt.Println(err)
	}
}
