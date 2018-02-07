package fileutils

// WriterCloser is like an io.Discard but it implements io.CLoser too.
type WriterCloser struct {
}

func (*WriterCloser) Write(p []byte) (int, error) { return 0, nil }

func (*WriterCloser) Close() error { return nil }
