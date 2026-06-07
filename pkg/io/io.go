package iox

import "io"

// Reader defines the interface for reading operations
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer defines the interface for writing operations
type Writer interface {
	Write(p []byte) (n int, err error)
}

// ReadWriter combines Reader and Writer interfaces
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser combines Reader with Close
type ReadCloser interface {
	Reader
	io.Closer
}

// WriteCloser combines Writer with Close
type WriteCloser interface {
	Writer
	io.Closer
}

// ReadWriteCloser combines Reader, Writer, and Close
type ReadWriteCloser interface {
	Reader
	Writer
	io.Closer
}

// Closer defines the interface for closing operations
type Closer interface {
	Close() error
}

// Seeker defines the interface for seeking operations
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadSeeker combines Reader and Seeker
type ReadSeeker interface {
	Reader
	Seeker
}

// WriteSeeker combines Writer and Seeker
type WriteSeeker interface {
	Writer
	Seeker
}

// ReadWriteSeeker combines Reader, Writer, and Seeker
type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// ReaderAt defines the interface for reading at offset
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// WriterAt defines the interface for writing at offset
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// ByteReader defines the interface for reading single bytes
type ByteReader interface {
	ReadByte() (byte, error)
}

// ByteWriter defines the interface for writing single bytes
type ByteWriter interface {
	WriteByte(c byte) error
}

// ByteScanner combines ByteReader with UnreadByte
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

// RuneReader defines the interface for reading runes
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// RuneScanner combines RuneReader with UnreadRune
type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

// StringWriter defines the interface for writing strings
type StringWriter interface {
	WriteString(s string) (n int, err error)
}

// ReaderFrom defines the interface for reading from a Reader
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo defines the interface for writing to a Writer
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

// Made with Bob
