package mapper

import (
	"encoding/json"
	"errors"
)

func ToJson[anything any](a anything) (string, error) {
	str, err := json.Marshal(a)
	return string(str), err
}

func FromJson[anything any](jsoned string) (anything, error) {
	str := []byte(jsoned)
	var a anything
	err := json.Unmarshal(str, &a)
	return a, errors.Wrap(err, "json.Unmarshal")
}
