package main

import (
	"github.com/Lobaro/slip"
	"bytes"
)

func main() {
	ReadExample()
	WriteExample()
}

func ReadExample() {
	data := []byte{1, 2, 3, slip.END}
	reader := slip.NewReader(bytes.NewReader(data))
	packet, isPrefix, err := reader.ReadPacket()

	// packet == 1, 2, 3
	// isPrefix == false
	// err == io.EOF

	if packet[0] != 1 || packet[1] != 2 || packet[2] != 3 {
		panic("Bad data")
	}
	if isPrefix {
		panic("isPrefix != false")
	}
	if err != nil {
		panic(err)
	}
}

func WriteExample() {
	buf := &bytes.Buffer{}
	writer := slip.NewWriter(buf)
	err := writer.WritePacket([]byte{1, 2, 3})

	// buf.Bytes() ==  [END, 1, 2, 3, END]

	packet := buf.Bytes()
	if packet[0] != slip.END || packet[1] != 1 || packet[2] != 2 || packet[3] != 3 || packet[4] != slip.END {
		panic("Bad data")
	}
	if err != nil {
		panic(err)
	}
}
