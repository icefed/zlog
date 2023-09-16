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
	// truncate nagative
	l := b.Len()
	b.Truncate(-1)
	if b.Len() != l {
		t.Errorf("got %q, want %q", b.Len(), l)
	}
	// truncate grow
	b.Truncate(100)
	if b.Len() != 100 {
		t.Errorf("got %q, want %q", b.Len(), 100)
	}
	// truncate shrink
	truncatedSize := 9
	truncatedBytes := "string by"
	b.Truncate(truncatedSize)
	if b.String() != truncatedBytes {
		t.Errorf("got %q, want %q", b.String(), truncatedBytes)
	}
	if b.Len() != truncatedSize {
		t.Errorf("got %q, want %q", b.Len(), truncatedSize)
	}
	// grow negative
	l = b.Len()
	b.Grow(-1)
	if b.Len() != l {
		t.Errorf("got %q, want %q", b.Len(), l)
	}
	// grow
	growSize := 10
	b.Grow(growSize)
	if b.Len() != l+growSize {
		t.Errorf("got %q, want %q", b.Len(), l+growSize)
	}
	l = b.Len()
	growSize = 500
	b.Grow(growSize)
	if b.Len() != l+growSize {
		t.Errorf("got %q, want %q", b.Len(), l+growSize)
	}
	l = b.Len()
	growSize = 1000
	b.Grow(growSize)
	if b.Len() != l+growSize {
		t.Errorf("got %q, want %q", b.Len(), l+growSize)
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
