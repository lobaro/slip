package slip

import (
	"fmt"
	"testing"
)

func TestCoapFcs(t *testing.T) {
	// Last 2 bytes are the CRC
	packet := []byte{100, 69, 0, 1, 1, 186, 27, 2, 255, 43, 228, 64, 44}

	if !CheckFsc16(packet) {
		t.Error("Bad CRC")
	}

	expected := CalcFcs16(packet)

	if expected != FCS_GOOD {
		t.Errorf("Expected FCS %02x but got %02x", FCS_GOOD, expected)
	}
}

func TestFcs(t *testing.T) {
	cp := []byte{1, 2, 3, 4, 5}

	trialfcs := CalcFcs16WithInit(FCS_INITIAL, cp)
	trialfcs ^= 0xffff                             // complement
	cp = append(cp, byte(trialfcs&uint16(0x00ff))) // least significant byte first
	cp = append(cp, byte((trialfcs>>8)&uint16(0x00ff)))

	t.Log(cp)

	/* check on input */
	trialfcs = CalcFcs16WithInit(FCS_INITIAL, cp)
	if trialfcs != 0xf0b8 {
		t.Error(fmt.Sprintf("Bad FCS %#04x, expected 0xf0b8", trialfcs))
	}
}

func TestFcsApi(t *testing.T) {
	message := []byte("Hallo!")

	if CheckFsc16(message) {
		t.Error("Expected Bad FCS")
	}
	fcs := CalcFcs16(message)

	packet := AppendFcs16(message, fcs)

	if !CheckFsc16(packet) {
		t.Error("Bad FCS in ", packet)
	}

}
