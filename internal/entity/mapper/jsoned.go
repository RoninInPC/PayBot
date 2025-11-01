package mapper

import "encoding/json"

func ToJson[anything any](a anything) (string, error) {
	str, err := json.Marshal(a)
	return string(str), err
}

func FromJson[anything any](jsoned string) (anything, error) {
	str := []byte(jsoned)
	var a anything
	err := json.Unmarshal(str, &a)
	return a, err
}
