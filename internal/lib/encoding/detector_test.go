package encoding

import (
	"bytes"
	"testing"
)

func TestDetectAndConvert_UTF8(t *testing.T) {
	input := []byte("Data,Valor,Descrição\n01/01/2026,100.00,Teste")
	result, err := DetectAndConvert(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := result.String()
	if !bytes.Contains([]byte(output), []byte("Descrição")) {
		t.Errorf("expected UTF-8 preserved, got: %s", output)
	}
}

func TestDetectAndConvert_ISO88591(t *testing.T) {
	// ISO-8859-1 encoded "Lançamento"
	input := []byte{0x4C, 0x61, 0x6E, 0xE7, 0x61, 0x6D, 0x65, 0x6E, 0x74, 0x6F}
	result, err := DetectAndConvert(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := result.String()
	if !bytes.Contains([]byte(output), []byte("Lançamento")) {
		t.Errorf("expected conversion to UTF-8, got: %s", output)
	}
}

func TestDetectAndConvert_Windows1252(t *testing.T) {
	// Windows-1252 encoded "Cartão"
	input := []byte{0x43, 0x61, 0x72, 0x74, 0xE3, 0x6F}
	result, err := DetectAndConvert(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := result.String()
	if !bytes.Contains([]byte(output), []byte("Cartão")) {
		t.Errorf("expected conversion to UTF-8, got: %s", output)
	}
}
