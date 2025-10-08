package netio

import (
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// IsBenign returns true if the error is a normal network closure
func IsBenign(err error) bool {
	if err == nil || err == io.EOF {
		return true
	}

	if ne, ok := err.(*net.OpError); ok && ne.Err != nil {
		s := ne.Err.Error()
		switch {
		case strings.Contains(s, "use of closed network connection"),
			strings.Contains(s, "connection reset by peer"),
			strings.Contains(s, "broken pipe"),
			strings.Contains(s, "i/o timeout"):
			return true
		}
	}

	if se, ok := err.(*os.SyscallError); ok && se.Err != nil {
		s := se.Err.Error()
		switch {
		case strings.Contains(s, "connection reset by peer"),
			strings.Contains(s, "broken pipe"),
			strings.Contains(s, "i/o timeout"):
			return true
		}
	}

	if pe, ok := err.(*os.PathError); ok && pe.Err != nil {
		if strings.Contains(pe.Err.Error(), "file already closed") {
			return true
		}
	}

	return false
}

// CopyPipe copies from src to dst and logs unexpected errors
func CopyPipe(dst io.WriteCloser, src io.ReadCloser, context string, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := io.Copy(dst, src)
	if !IsBenign(err) {
		log.Printf("%s: %v", context, err)
	}
	_ = dst.Close()
}

// BiCopy starts bidirectional copying between two connections
// and waits until both directions finish.
func BiCopy(a io.ReadWriteCloser, b io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(2)
	go CopyPipe(b, a, "a->b", &wg)
	go CopyPipe(a, b, "b->a", &wg)
	wg.Wait()
}
