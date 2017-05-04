package slip

import (
	"testing"
	"bytes"
	"io"
	"strconv"
)

var readData = []struct {
	data     []byte
	expected []byte
	isPrefix bool
	err      error
}{
	{[]byte{}, []byte{}, true, io.EOF},
	// All the END are received till EOF or data
	{[]byte{END, END, END, END}, []byte{}, true, io.EOF},
	{[]byte{END, END, 1, END}, []byte{1}, false, nil},
	// Properly terminated data
	{[]byte{1, 2, 3, END}, []byte{1, 2, 3}, false, nil},
	{[]byte{1, 2, 3, END, 4, 5, 6}, []byte{1, 2, 3}, false, nil},
	{[]byte{ESC, ESC_ESC, END}, []byte{ESC}, false, nil},
	{[]byte{ESC, ESC_END, END}, []byte{END}, false, nil},
	// Non terminated data
	{[]byte{1, 2, 3}, []byte{1, 2, 3}, true, io.EOF},
	{[]byte{ESC, ESC_ESC}, []byte{ESC}, true, io.EOF},
	{[]byte{ESC, ESC_END}, []byte{END}, true, io.EOF},
	// Bad control sequences
	{[]byte{1, ESC_ESC, 3}, []byte{1, ESC_ESC, 3}, true, io.EOF},
	{[]byte{1, ESC_END, 3}, []byte{1, ESC_END, 3}, true, io.EOF},
	{[]byte{1, ESC, 3}, []byte{1, 3}, true, io.EOF},
}

var writeData = []struct {
	data     []byte
	expected []byte
	err      error
}{
	// Just data. Starts with END and ends with END
	{[]byte{1, 2, 3}, []byte{END, 1, 2, 3, END}, nil},
	// Escape sequences
	{[]byte{END}, []byte{END, ESC, ESC_END, END}, nil},
	{[]byte{ESC}, []byte{END, ESC, ESC_ESC, END}, nil},
	{[]byte{ESC_END}, []byte{END, ESC_END, END}, nil},
	{[]byte{ESC_ESC}, []byte{END, ESC_ESC, END}, nil},
	{[]byte{END, ESC}, []byte{END, ESC, ESC_END, ESC, ESC_ESC, END}, nil},
}

func eqBytes(a, b []byte) bool {
	if len(a) != len(b) {

		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRead(t *testing.T) {
	for i, d := range readData {
		r := NewReader(bytes.NewReader(d.data))
		p, isPrefix, err := r.ReadPacket()

		if err == nil && d.err != nil {
			t.Error(strconv.Itoa(i), "Expected error", d.err.Error(), "but got", err)
		} else if err != nil && d.err == nil {
			t.Error(strconv.Itoa(i), "Expected error", d.err, "but got", err)
		} else if err != d.err && err.Error() != d.err.Error() {
			t.Error(strconv.Itoa(i), "Expected error", d.err, "but got", err)
		}

		if isPrefix != d.isPrefix {
			t.Error(strconv.Itoa(i), "Expected isPrefix", d.isPrefix, "but got", isPrefix)
		}
		if !eqBytes(p, d.expected) {
			t.Error(strconv.Itoa(i), "Expected data", d.expected, "but got", p)
		}
	}
}

func TestWrite(t *testing.T) {
	for i, d := range writeData {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WritePacket(d.data)

		if err == nil && d.err != nil {
			t.Error(strconv.Itoa(i), "Expected error", d.err.Error(), "but got", err)
		} else if err != nil && d.err == nil {
			t.Error(strconv.Itoa(i), "Expected error", d.err, "but got", err)
		} else if err != d.err && err.Error() != d.err.Error() {
			t.Error(strconv.Itoa(i), "Expected error", d.err, "but got", err)
		}

		if !eqBytes(buf.Bytes(), d.expected) {
			t.Error(strconv.Itoa(i), "Expected data", d.expected, "but got", buf.Bytes())
		}
	}
}

func TestWriteAndRead(t *testing.T) {
	for i, d := range writeData {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WritePacket(d.data)

		if err != nil {
			t.Error("Unexpected error:", err)
		}

		r := NewReader(buf)
		p, isPrefix, err := r.ReadPacket()

		if err != nil {
			t.Error("Unexpected error:", err)
		}
		if isPrefix {
			t.Error("Expected no Prefix but is", isPrefix)
		}
		if !eqBytes(p, d.data) {
			t.Error(strconv.Itoa(i), "Expected data", d.data, "but got", p)
		}
	}
}


// Encoded holds data that can be retrived via all interfaces
// defined in the "encoding" package
type EncodedBin struct {
	data []byte
}

func (e *EncodedBin) MarshalBinary() (data []byte, err error) {
	return e.data, nil
}
func (e *EncodedBin) UnmarshalBinary(data []byte) error {
	e.data = data
	return nil
}

// Encoded holds data that can be retrived via all interfaces
// defined in the "encoding" package
type EncodedText struct {
	data []byte
}

func (e *EncodedText) MarshalText() (text []byte, err error) {
	return e.data, nil
}
func (e *EncodedText) UnmarshalText(text []byte) error {
	e.data = text
	return nil
}

// Test the interfaces from the encoding package
func TestEncodingInterfaces(t *testing.T) {
	encBin := EncodedBin{data: []byte{1}}
	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)
	err := enc.Encode(encBin)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{END, 1, END}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Bytes not equal", expected, buf.Bytes())
	}

	encText := EncodedText{data: []byte{2}}
	buf = &bytes.Buffer{}
	enc = NewEncoder(buf)
	err = enc.Encode(encText)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{END, 2, END}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Bytes not equal", expected, buf.Bytes())
	}

	buf = &bytes.Buffer{}
	enc = NewEncoder(buf)
	err = enc.Encode(byte(3))
	if err != nil {
		t.Error(err)
	}
	expected = []byte{END, 3, END}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Bytes not equal", expected, buf.Bytes())
	}

	buf = &bytes.Buffer{}
	enc = NewEncoder(buf)
	err = enc.Encode([]byte{4})
	if err != nil {
		t.Error(err)
	}
	expected = []byte{END, 4, END}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Error("Bytes not equal", expected, buf.Bytes())
	}
}

// Test the interfaces from the encoding package
func TestDecodingInterfaces(t *testing.T) {
	encBin := EncodedBin{}
	buf := bytes.NewBuffer([]byte{END, 1, END})
	enc := NewDecoder(buf)
	err := enc.Decode(&encBin)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{1}
	if !bytes.Equal(encBin.data, expected) {
		t.Error("Bytes not equal", expected, encBin.data)
	}

	encText := EncodedText{}
	buf = bytes.NewBuffer([]byte{END, 2, END})
	enc = NewDecoder(buf)
	err = enc.Decode(&encText)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{2}
	if !bytes.Equal(encText.data, expected) {
		t.Error("Bytes not equal", expected, encText.data)
	}

	target := []byte{}
	buf = bytes.NewBuffer([]byte{END, 3, END})
	enc = NewDecoder(buf)
	err = enc.Decode(&target)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{3}
	if !bytes.Equal(target, expected) {
		t.Error("Bytes not equal", expected, target)
	}

	target = []byte{}
	buf = bytes.NewBuffer([]byte{END, 4, END})
	enc = NewDecoder(buf)
	err = enc.Decode([]byte{END, 4, END})
	if err != nil {
		t.Error(err)
	}
	expected = []byte{4}
	if !bytes.Equal(target, expected) {
		t.Error("Bytes not equal", expected, target)
	}
}