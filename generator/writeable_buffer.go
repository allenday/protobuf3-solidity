package generator

import (
	"bytes"
	"fmt"
)

// WriteableBuffer is a buffer that can be written to with indentation
type WriteableBuffer struct {
	buffer bytes.Buffer
	indent string
}

// NewWriteableBuffer creates a new WriteableBuffer
func NewWriteableBuffer() *WriteableBuffer {
	return &WriteableBuffer{
		indent: "",
	}
}

// P prints a line to the buffer with indentation
func (b *WriteableBuffer) P(format ...interface{}) {
	if len(format) > 0 {
		b.buffer.WriteString(b.indent)
		if len(format) == 1 {
			if str, ok := format[0].(string); ok {
				b.buffer.WriteString(str)
			} else {
				fmt.Fprintf(&b.buffer, "%v", format[0])
			}
		} else {
			// Extract format string and arguments properly
			formatStr, ok := format[0].(string)
			if !ok {
				// Fallback: just print all arguments as is
				for i, arg := range format {
					if i > 0 {
						b.buffer.WriteByte(' ')
					}
					fmt.Fprintf(&b.buffer, "%v", arg)
				}
			} else {
				// Use fmt.Fprintf directly with the format string and arguments
				fmt.Fprintf(&b.buffer, formatStr, format[1:]...)
			}
		}
	}
	b.buffer.WriteByte('\n')
}

// Indent increases the indentation level
func (b *WriteableBuffer) Indent() {
	b.indent += "\t"
}

// Unindent decreases the indentation level
func (b *WriteableBuffer) Unindent() {
	if len(b.indent) > 0 {
		b.indent = b.indent[1:]
	}
}

// String returns the contents of the buffer as a string
func (b *WriteableBuffer) String() string {
	return b.buffer.String()
}

// Bytes returns the contents of the buffer as a byte slice
func (b *WriteableBuffer) Bytes() []byte {
	return b.buffer.Bytes()
}

// Reset resets the buffer
func (b *WriteableBuffer) Reset() {
	b.buffer.Reset()
	b.indent = ""
}

// P0 prints an empty line to the buffer
func (b *WriteableBuffer) P0() {
	b.buffer.WriteByte('\n')
}
