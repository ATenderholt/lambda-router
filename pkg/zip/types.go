package zip

import "io"

type ZipContent struct {
	Offset  int64
	Content []byte
	Length  int64
}

type ZipFileError struct {
	Message  string
	Filepath string
	Err      error
}

func (e ZipFileError) Error() string {
	return e.Message + " " + e.Filepath + " : " + e.Err.Error()
}

func min(a int, b int64) int64 {
	a64 := int64(a)
	if a64 < b {
		return a64
	}

	return b
}

func (source ZipContent) ReadAt(p []byte, off int64) (n int, err error) {
	logger.Debug("Attempting to read %d bytes from offset %d", len(p), off)

	if off >= source.Length {
		return 0, io.EOF
	}

	bytesToRead := min(len(p), source.Length-off)
	count := copy(p, source.Content[off:off+bytesToRead])

	if count < len(p) {
		return count, io.EOF
	}

	return count, nil
}
