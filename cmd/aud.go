package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/siggy/bbox/bbox"
)

func main() {
	files, _ := ioutil.ReadDir(bbox.WAVS)
	if len(files) != bbox.SOUNDS {
		panic(0)
	}

	wavs := bbox.InitWavs()

	for i := 0; i < bbox.SOUNDS; i++ {
		fmt.Printf("Play(%+v)\n", i)
		wavs.Play(i)
		time.Sleep(2 * time.Second)
		// break
	}

	wavs.Close()
}
