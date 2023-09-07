package buffer

import "testing"

func Test(t *testing.T) {
	b := New()
	defer b.Free()
	b.WriteString("string")
	b.WriteByte(' ')
	b.Write([]byte("bytes"))
	b.WriteRune('😀')
	b.WriteByte('\n')

	wantBytes := "string bytes😀\n"
	if string(b.Bytes()) != wantBytes {
		t.Errorf("got %q, want %q", b.Bytes(), wantBytes)
	}
	if b.Len() != len(wantBytes) {
		t.Errorf("got %q, want %q", b.Len(), len(wantBytes))
	}
	// last byte
	var wantLastByte byte = '\n'
	if *b.LastByte() != wantLastByte {
		t.Errorf("got %q, want %q", *b.LastByte(), len(wantBytes))
	}
	// truncate
	b.Truncate(100)
	if string(b.Bytes()) != wantBytes {
		t.Errorf("got %q, want %q", b.Bytes(), wantBytes)
	}
	truncatedSize := 9
	truncatedBytes := "string by"
	b.Truncate(truncatedSize)
	if string(b.Bytes()) != truncatedBytes {
		t.Errorf("got %q, want %q", b.Bytes(), truncatedBytes)
	}
	if b.Len() != truncatedSize {
		t.Errorf("got %q, want %q", b.Len(), truncatedSize)
	}

	// reset
	b.Reset()
	if b.Len() != 0 {
		t.Errorf("got %q, want %q", b.Len(), 0)
	}
	// last byte
	if b.LastByte() != nil {
		t.Errorf("got %q, want %v", *b.LastByte(), nil)
	}
}
