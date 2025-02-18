package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	StyleReset   = "\033[0m"
	StyleBold    = "\033[1m"
	StyleFgGreen = "\033[32m"
)

type PrefixedWriter struct {
	w      io.Writer
	prefix []byte
	buf    *bytes.Buffer
}

func (pw *PrefixedWriter) Write(p []byte) (int, error) {
	defer pw.buf.Reset()
	n, err := pw.buf.Write(p)
	if err != nil {
		return n, err
	}

	for {
		line, err := pw.buf.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			pw.buf.Reset()
			pw.buf.Write(line)
			break
		}

		if _, err := pw.w.Write(append(pw.prefix, line...)); err != nil {
			return n, err
		}
	}
	return n, nil
}

var _ io.Writer = (*PrefixedWriter)(nil)

type LogWriter struct {
	w  io.Writer
	mu sync.Mutex
	wg sync.WaitGroup
}

// Write implements io.Writer.
func (s *LogWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

var _ io.Writer = (*LogWriter)(nil)

func (s *LogWriter) WithPrefix(prefix string) io.Writer {
	if prefix != "" && hasANSISupport() {
		prefix = fmt.Sprintf("%s[%s]%s ", StyleFgGreen, prefix, StyleReset)
		// prefix = fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
	}

	return &PrefixedWriter{s.w, []byte(prefix), bytes.NewBuffer(nil)}
}
