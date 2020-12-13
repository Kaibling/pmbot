package utils

import "encoding/json"

//ChannelMessage -
type ChannelMessage struct {
	Topic   string
	Content interface{}
}

func pretty(data interface{}) string {
	a, _ := json.MarshalIndent(data, "", " ")
	return string(a)
}
