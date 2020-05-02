package compression

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func createResponseWriter(config *Config, w http.ResponseWriter, r *http.Request) (*responseWriter, bool) {
	encoding := negotiateContentEncoding(r, config.supportedEncodings)
	compressor := config.Compressors.Find(encoding)
	if compressor == nil {
		return nil, false
	}

	proxyWriter := &responseWriter{
		ResponseWriter: w,
		compressor:     compressor,

		status: 200, // default
	}
	return proxyWriter, true
}

// copied from gddo
// but without parsing Q value
func negotiateContentEncoding(r *http.Request, offers []string) string {
	bestOffer := ""
	specs := r.Header.Values("accept-encoding")
	for _, offer := range offers {
		for _, spec := range specs {
			splitted := strings.SplitN(spec, ";", 2)
			value := strings.TrimSpace(splitted[0])

			if value == "*" || value == offer {
				bestOffer = offer
			}
		}
	}
	return bestOffer
}

type responseWriter struct {
	// required fields
	http.ResponseWriter
	compressor Compressor

	minCompressLength int

	// must be initialized
	status int

	// internal use
	err     error
	written bool

	buffer        *bytes.Buffer
	length        int
	writeToBuffer bool

	currentWriter io.Writer
	currentCloser io.Closer
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	if w.err != nil {
		return 0, w.err
	}

	w.length += len(data)

	if !w.writeToBuffer {
		// we already know the size, write directly to writer
		return w.currentWriter.Write(data)
	}

	if w.length >= w.minCompressLength {
		// stop writing to buffer and start the compression
		w.writeToBuffer = false
		err := w.startCompress()
		if err != nil {
			w.err = err
			return 0, err
		}
		// flush buffer content to writer
		_, err = w.buffer.WriteTo(w.currentWriter)
		if err != nil {
			return 0, err
		}

		// write current data
		return w.currentWriter.Write(data)
	}

	// append data to buffer
	return w.buffer.Write(data)
}

func (w *responseWriter) Size() int {
	return w.length
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) WriteHeaderNow() {
	if !w.written {
		w.written = true

		knownLength := w.getContentLength()
		if w.Header().Get("content-encoding") != "" || (knownLength >= 0 && knownLength < w.minCompressLength) {
			// if it already compressed, leave it as it's
			// or if we know the size is less than minCompressLength
			w.ResponseWriter.WriteHeader(w.status)
			w.currentWriter = w.ResponseWriter
		} else if knownLength > w.minCompressLength {
			// okay, we know the length and it's big enough
			err := w.startCompress()
			if err != nil {
				w.err = err
				return
			}
		} else {
			// we don't know the actual length
			// write to buffer first
			w.buffer = &bytes.Buffer{}
			w.writeToBuffer = true
		}
	}
}

func (w *responseWriter) startCompress() error {
	// we don't know the compressed size
	// remove content-length header
	w.Header().Del("content-length")
	w.Header().Set("content-encoding", w.compressor.Encoding())
	w.ResponseWriter.WriteHeader(w.status)

	writer, err := w.compressor.NewWriter(w.ResponseWriter)
	if err != nil {
		return err
	}
	w.currentWriter = writer
	w.currentCloser = writer
	return nil
}

func (w *responseWriter) end() {
	if w.currentCloser != nil {
		w.currentCloser.Close()
	}

	if w.err != nil {
		http.Error(w.ResponseWriter, "Compression error: "+w.err.Error(), http.StatusInternalServerError)
		return
	}

	if w.writeToBuffer {
		// we know the length because it's not compressed
		w.Header().Set("content-length", strconv.Itoa(w.length))
		w.ResponseWriter.WriteHeader(w.status)
		w.buffer.WriteTo(w.ResponseWriter)
	}
}

func (w *responseWriter) getContentLength() int {
	if lengthStr := w.Header().Get("content-length"); lengthStr != "" {
		length, err := strconv.ParseInt(lengthStr, 10, 32)
		if err == nil {
			return int(length)
		}
	}
	return -1
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
}
func (w *responseWriter) Written() bool {
	return w.written
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {

	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (w *responseWriter) Flush() {
	w.WriteHeaderNow()
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *responseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
