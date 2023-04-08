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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	FfmpegExecutable   = "./ffmpeg"
	CwebpExecutable    = "./cwebp"
	Gif2webpExecutable = "./gif2webp"
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

	// Need to do this in a go routine because dapr issues a timeout after a certain amount of time. This seems not to be configurable in the open function framework
	go func() {
		err = MediaConversion(mediaEvent, logger)
		if err != nil {
			logger.Error("error converting", zap.Error(err))
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

func MediaConversion(mediaEvent MediaEvent, logger *zap.Logger) error {
	logger.Debug("ffmpeg executable path", zap.String("path", FfmpegExecutable))
	logger.Debug("cwebp executable path", zap.String("path", CwebpExecutable))
	logger.Debug("gif2webp executable path", zap.String("path", Gif2webpExecutable))

	id, extension := getMediaIdComponents(mediaEvent.MediaId)

	// TODO: Could do tracing here with traceparent
	response, err := http.Get(fmt.Sprintf("http://localhost:8083/media/%s", mediaEvent.MediaId))
	if err != nil {
		logger.Error("error getting presigned url", zap.Error(err))
		return err
	}

	err = writeToDisk(mediaEvent.MediaId, response.Body)
	if err != nil {
		logger.Error("error writing to disk", zap.Error(err))
		return err
	}

	done := make(chan FileOp, 1)
	go func() {
		switch extension {
		case "mp4":
			done <- mp4ToWebm(id, logger)
			break
		case "png", "jpeg", "gif":
			done <- imgToWebp(id, extension, logger)
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
		return err
	}

	logger.Debug("Waiting for conversion")
	fileOp := <-done
	if fileOp.err != nil {
		logger.Debug("error while converting", zap.Error(err))
		return fileOp.err
	}
	logger.Debug("Finished conversion", zap.String("outputPath", fileOp.path))

	err = upload(client, fileOp)
	if err != nil {
		logger.Error("error uploading", zap.Error(err))
		return err
	}

	err = cleanUp(mediaEvent.MediaId, fileOp.path)
	if err != nil {
		logger.Error("error cleaning up", zap.Error(err))
		return err
	}
	return nil
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

func imgToWebp(id string, extension string, logger *zap.Logger) FileOp {
	input := fmt.Sprintf("%s.%s", id, extension)
	output := fmt.Sprintf("%s.webp", id)

	outputFile, err := os.Create(output)
	if err != nil {
		return FileOp{err: err}
	}
	_ = outputFile.Close()

	var stdOut []byte
	switch extension {
	case "gif":
		stdOut, err = exec.Command(Gif2webpExecutable, input, "-o", output).CombinedOutput()
		break
	case "png", "jpeg":
		stdOut, err = exec.Command(CwebpExecutable, input, "-o", output).CombinedOutput()
		break
	default:
		err = errors.New("unrecognized format")
		stdOut = make([]byte, 0)
	}

	logger.Debug("webp output", zap.ByteString("output", stdOut))

	if err != nil {
		return FileOp{err: err}
	}

	return FileOp{path: output}
}

func mp4ToWebm(id string, logger *zap.Logger) FileOp {
	buf := &bytes.Buffer{}

	input := fmt.Sprintf("%s.mp4", id)
	output := fmt.Sprintf("%s.webm", id)

	err := fluentffmpeg.NewCommand(FfmpegExecutable).
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

	out, _ := io.ReadAll(buf)
	logger.Debug("ffmpeg output", zap.String("output", string(out)))

	if err != nil {
		return FileOp{err: err}
	}

	return FileOp{path: output}
}
