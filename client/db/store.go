package db

type Store interface {
	Close(path string)
	Get(path string) []byte
	Post(path string, nfo interface{}) (err error)
	Put(path string, offset int64)
	IsDone(path string) int64
	Delete(path string)
}
