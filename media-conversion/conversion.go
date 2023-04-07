package media_conversion

import (
	"context"
	"encoding/json"
	"fmt"
	ofctx "github.com/OpenFunction/functions-framework-go/context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
)

func HandleMessage(ofctx ofctx.Context, in []byte) (ofctx.Out, error) {
	var mediaEvent MediaEvent
	err := json.Unmarshal(in, &mediaEvent)
	if err != nil {
		fmt.Println("error reading message from Redis binding", err)
		return ofctx.ReturnOnInternalError(), err
	}
	fmt.Printf("message from Redis '%s'\n", mediaEvent)

	// TODO: Could do tracing here with traceparent
	resp, err := http.Get(fmt.Sprintf("http://localhost:8083/media/%s", mediaEvent.MediaId))
	if err != nil {
		fmt.Println("error getting presigned url", err)
		return ofctx.ReturnOnInternalError(), err
	}

	bodyReader := resp.Body

	decode, err := png.Decode(bodyReader)
	if err != nil {
		fmt.Println("error decoding", err)
		return ofctx.ReturnOnInternalError(), err
	}

	// Create a new grayscale image
	bounds := decode.Bounds()
	grayImage := image.NewGray(bounds)

	// Convert each pixel to grayscale
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			oldColor := decode.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor).(color.Gray)
			grayImage.Set(x, y, grayColor)
		}
	}

	// TODO: Shouldnt hardcode
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		fmt.Println("error creating minio client", err)
		return ofctx.ReturnOnInternalError(), err
	}

	reader, writer := io.Pipe()
	go func() {
		if err := png.Encode(writer, grayImage); err != nil {
			fmt.Println("Error:", err)
			return
		}
		err := writer.Close()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}()

	slices := strings.Split(mediaEvent.MediaId, ".")
	key := fmt.Sprintf("%s-gray.%s", slices[0], slices[1])
	ctx := context.Background()
	_, err = client.PutObject(ctx, "testbucket", key, reader, -1, minio.PutObjectOptions{})
	if err != nil {
		fmt.Println("Error:", err)
		return ofctx.ReturnOnInternalError(), err
	}

	return ofctx.ReturnOnSuccess(), nil
}

type MediaEvent struct {
	MediaId string
}
