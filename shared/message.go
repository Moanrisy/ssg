package shared

type ContentType int

const (
	MESSAGE ContentType = iota
	INPUT
)

type Message struct {
	Content string
	Type    ContentType
}
