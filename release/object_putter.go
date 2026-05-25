package release

type ObjectPutter interface {
	Put(key string, body []byte, contentType string) error
}
