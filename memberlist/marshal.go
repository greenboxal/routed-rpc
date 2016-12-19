package memberlist

import (
	"bytes"
	"encoding/gob"
)

func encode(value interface{}) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(buff).Encode(&value)

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func decode(data []byte) (interface{}, error) {
	var value interface{}

	buff := bytes.NewBuffer(data)

	err := gob.NewDecoder(buff).Decode(&value)

	if err != nil {
		return nil, err
	}

	return value, nil
}
