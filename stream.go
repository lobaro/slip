package slip

import (
	"encoding"
	"errors"
	"io"
	"reflect"
)

// EncodedLen returns the length of an encoding of n source bytes.
func EncodedLen(n int) int {
	// TODO: Count bytes to be escaped!
	return n + 2
}

func DecodedLen(x int) int {
	// TODO: Count byytes that will be dropped!
	return x - 2
}

// A Decoder reads and decodes SLIP values from an input stream.
type Decoder struct {
	r SlipReader
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the SLIP packet requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: NewReader(r)}
}

// Decode reads the next SLIP-encoded packet from its
// input and stores it in the value pointed to by v.
func (d *Decoder) Decode(v interface{}) error {
	buf := make([]byte, 0)
	for {
		part, isPrefix, err := d.r.ReadPacket()
		buf := append(buf, part...)
		if err != nil {
			return err
		}
		if !isPrefix {
			return toStruct(buf, v)
		}
	}

	return nil
}

func toStruct(data []byte, val interface{}) error {

	switch v := val.(type) {
	case encoding.BinaryUnmarshaler:
		return v.UnmarshalBinary(data)
	case encoding.TextUnmarshaler:
		return v.UnmarshalText(data)
	case []byte:
		val = v
		return nil
	case byte:
		val = []byte{v}
		return nil
	case string:
		val = []byte(v)
		return nil
	}
	return errors.New("slip: Failed to unmarshal value of type " + reflect.TypeOf(val).String())
}

// An Encoder writes SLIP packets to an output stream.
type Encoder struct {
	w SlipWriter
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: NewWriter(w)}
}

// Encode writes the SLIP encoding of v to the stream.
// Starting and ending with an END byte
//
// Values implementing encoding.BinaryMarshaler or
// encoding.TextMarshaler are marshaled using the Marshal* functions.
// Slices/arrays of bytes, single bytes, and strings are just
// converted to []byte without touching the content.
func (e *Encoder) Encode(v interface{}) error {
	data, err := toBytes(v)
	if err != nil {
		return err
	}
	return e.w.WritePacket(data)
}

func toBytes(v interface{}) ([]byte, error) {
	if v, ok := v.(encoding.BinaryMarshaler); ok {
		return v.MarshalBinary()
	}
	if v, ok := v.(encoding.TextMarshaler); ok {
		text, err := v.MarshalText()
		return []byte(text), err
	}

	switch v := v.(type) {
	case []byte:
		return v, nil
	case byte:
		return []byte{v}, nil
	case string:
		return []byte(v), nil
	}
	return nil, errors.New("slip: Failed to marshal value of type " + reflect.TypeOf(v).String())
}
