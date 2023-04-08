package main

import (
	"go.uber.org/zap"
	"sync"
	media_conversion "yatc.com/media-conversion"
)

func main() {
	logger, _ := zap.NewDevelopment()
	event := media_conversion.MediaEvent{MediaId: "a22e7ac7-d975-4361-aedf-7547fc978746.png"}
	//event := MediaEvent{MediaId: "bdce9518-44da-46f0-8837-e40f0cd2336d.jpeg"}
	//event := MediaEvent{MediaId: "22400f9c-e651-462d-b717-4b9f9dee89f6.mp4"}
	//event := MediaEvent{MediaId: "820f0954-9eb2-477e-8cd1-0c36bc320b37.gif"}

	media_conversion.CwebpExecutable = "cwebp"
	media_conversion.FfmpegExecutable = "ffmpeg"
	media_conversion.Gif2webpExecutable = "gif2webp"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := media_conversion.MediaConversion(event, logger)
		if err != nil {
			logger.Error("error converting", zap.Error(err))
			return
		}
		wg.Done()
	}()
	wg.Wait()
}
