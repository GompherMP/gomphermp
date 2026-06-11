package main

import (
	"fmt"
	"time"
)

func decodeVideo() {
	time.Sleep(40 * time.Millisecond)
	fmt.Println("video decoded")
}

func decodeAudio() {
	time.Sleep(20 * time.Millisecond)
	fmt.Println("audio decoded")
}

func loadSubtitles() {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("subtitles loaded")
}

func main() {
	//gompher parallel sections
	{
		//gompher section
		{
			decodeVideo()
		}

		//gompher section
		{
			decodeAudio()
		}

		//gompher section
		{
			loadSubtitles()
		}
	}

	fmt.Println("all media ready")
}
