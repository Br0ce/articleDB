package encoding

import (
	"bytes"
	"encoding/json"
	"io"
)

// DecodeJSON decodes the content of the reader into data.
func DecodeJSON(reader io.Reader, data interface{}) error {
	decoder := json.NewDecoder(reader)
	return decoder.Decode(data)
}

// EncodeJSON encodes data as json.
func EncodeJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// EncodeToReader encodes data as json and returns it as a reader.
func EncodeToReader(data interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// UnmashalJSON unmashals the given text into data.
func UnmashalJSON(text string, data interface{}) error {
	return json.Unmarshal([]byte(text), data)
}
