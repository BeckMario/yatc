package media_conversion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	ofctx "github.com/OpenFunction/functions-framework-go/context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	"github.com/nickalie/go-webpbin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func HandleMessage(ofctx ofctx.Context, in []byte) (ofctx.Out, error) {
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	var mediaEvent MediaEvent
	err := json.Unmarshal(in, &mediaEvent)
	if err != nil {
		logger.Error("error reading message from input binding", zap.Error(err))
		return ofctx.ReturnOnInternalError(), err
	}
	logger.Debug("received message from input binding", zap.String("MediaId", mediaEvent.MediaId))

	id, extension := getMediaIdComponents(mediaEvent.MediaId)

	// TODO: Could do tracing here with traceparent
	response, err := http.Get(fmt.Sprintf("http://localhost:8083/media/%s", mediaEvent.MediaId))
	if err != nil {
		logger.Error("error getting presigned url", zap.Error(err))
		return ofctx.ReturnOnInternalError(), err
	}

	err = writeToDisk(mediaEvent.MediaId, response.Body)
	if err != nil {
		logger.Error("error writing to disk", zap.Error(err))
		return ofctx.ReturnOnInternalError(), err
	}

	done := make(chan FileOp, 1)
	go func() {
		switch extension {
		case "mp4":
			done <- mp4ToWebm(id, logger)
			break
		case "png", "jpeg":
			done <- imgToWebp(id, extension)
			break
		case "gif":
			done <- gifToWebp(id)
			break
		default:
			done <- FileOp{err: errors.New("unrecognized format")}
		}
	}()

	// TODO: Shouldnt hardcode
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		logger.Error("error creating minio client", zap.Error(err))
		return ofctx.ReturnOnInternalError(), err
	}

	// Need to do this in a go routine because dapr issues a timeout after a certain amount of time. This seems not to be configurable in the open function framework
	go func() {
		logger.Debug("Waiting for conversion")

		fileOp := <-done

		logger.Debug("Finished conversion", zap.String("outputPath", fileOp.path))

		err := upload(client, fileOp)
		if err != nil {
			logger.Error("error uploading", zap.Error(err))
			return
		}

		err = cleanUp(mediaEvent.MediaId, fileOp.path)
		if err != nil {
			logger.Error("error cleaning up", zap.Error(err))
			return
		}
	}()

	return ofctx.ReturnOnSuccess(), nil
}

type FileOp struct {
	path string
	err  error
}

type MediaEvent struct {
	MediaId string
}

func getMediaIdComponents(mediaId string) (string, string) {
	slices := strings.Split(mediaId, ".")
	return slices[0], slices[1]
}

func cleanUp(input string, output string) error {
	err := os.Remove(output)
	if err != nil {
		return err
	}
	err = os.Remove(input)
	if err != nil {
		return err
	}

	return nil
}

func upload(client *minio.Client, fileOp FileOp) error {
	if fileOp.err != nil {
		return fileOp.err
	}

	_, err := client.FPutObject(context.Background(), "testbucket", fileOp.path, fileOp.path, minio.PutObjectOptions{}) //reader, -1, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func writeToDisk(path string, reader io.Reader) error {
	input, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(input, reader)
	if err != nil {
		return err
	}

	err = input.Close()
	if err != nil {
		return err
	}
	return nil
}

func getDecoder(extension string) (func(r io.Reader) (image.Image, error), error) {
	switch extension {
	case "jpeg":
		return jpeg.Decode, nil
	case "png":
		return png.Decode, nil
	default:
		return nil, errors.New("unrecognized format")
	}
}

func gifToWebp(id string) FileOp {
	input := fmt.Sprintf("%s.gif", id)
	output := fmt.Sprintf("%s.webp", id)

	inputFile, err := os.Open(input)
	if err != nil {
		return FileOp{err: err}
	}

	defer func(inputFile *os.File) {
		_ = inputFile.Close()
	}(inputFile)

	outputFile, err := os.Create(output)
	if err != nil {
		return FileOp{err: err}
	}

	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)

	err = exec.Command("./gif2webp", input, "-o", output).Run()
	if err != nil {
		return FileOp{err: err}
	}
	return FileOp{path: output}
}

func imgToWebp(id string, extension string) FileOp {
	input := fmt.Sprintf("%s.%s", id, extension)
	output := fmt.Sprintf("%s.webp", id)

	inputFile, err := os.Open(input)
	if err != nil {
		return FileOp{err: err}
	}

	defer func(inputFile *os.File) {
		_ = inputFile.Close()
	}(inputFile)

	decoder, err := getDecoder(extension)
	if err != nil {
		return FileOp{err: err}
	}

	img, err := decoder(inputFile)
	if err != nil {
		return FileOp{err: err}
	}

	outputFile, err := os.Create(output)
	if err != nil {
		return FileOp{err: err}
	}

	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)

	err = webpbin.Encode(outputFile, img)
	if err != nil {
		return FileOp{err: err}
	}

	return FileOp{path: output}
}

func mp4ToWebm(id string, logger *zap.Logger) FileOp {
	buf := &bytes.Buffer{}

	input := fmt.Sprintf("%s.mp4", id)
	output := fmt.Sprintf("%s.webm", id)

	err := fluentffmpeg.NewCommand("./ffmpeg").
		InputPath(input).
		AudioCodec("libopus").
		AudioBitRate(48000).
		VideoCodec("libvpx-vp9").
		VideoBitRate(0).
		ConstantRateFactor(50).
		OutputFormat("webm").
		OutputOptions("-deadline", "realtime", "-cpu-used", "-8", "-vf",
			"scale='min(1280,iw)':min'(720,ih)':force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2").
		OutputLogs(buf).
		OutputPath(output).
		Run()

	out, _ := io.ReadAll(buf) // read logs
	logger.Debug("ffmpeg output", zap.String("output", string(out)))

	if err != nil {
		return FileOp{err: err}
	}

	return FileOp{path: output}
}
