package main

import (
	"fmt"
	"github.com/hackborn/phly"
	_ "github.com/hackborn/phly_img"
)

func main() {
	_, err := phly.RunApp()
	if err != nil {
		fmt.Println(err)
	}
}
