package utils

import (
	"encoding/json"
)

func Pretty(data interface{}) string {
	a, _ := json.MarshalIndent(data, "", " ")
	return string(a)
}
