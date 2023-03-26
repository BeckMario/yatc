package media

import "io"

type Metadata struct {
	Format string
}

type Media struct {
	metadata Metadata
	fileName string
	reader   *io.ReadCloser
}

type Service interface {
	UploadFile(media *Media)
	DownloadFile()
}
