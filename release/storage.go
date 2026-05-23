package release

type Storage interface {
	UploadFile(from string, to string) error
	UploadContent(content string, to string) error
	DownloadContent(from string) (string, error)
}
