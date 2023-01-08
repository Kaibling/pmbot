package models

type Message struct {
	Topic   string
	Content interface{}
	Sender  string
}

func (m *Message) ContentStr() string {
	return m.Content.(string)
}
