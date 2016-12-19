package routedrpc

import (
	"bytes"
	"encoding/gob"
)

func marshal(obj interface{}) []byte {
	buff := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(buff).Encode(obj)

	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}

func unmarshal(data []byte) interface{} {
	var result interface{}

	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(result)

	if err != nil {
		panic(err)
	}

	return result
}
