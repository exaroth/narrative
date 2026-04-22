package reader

type SourceReader interface {
	Read(string) (SourceReader, error)
	Data() []byte
	Title() string
	Author() string
	Id() string
}
