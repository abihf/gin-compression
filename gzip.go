package compression

import (
	"compress/gzip"
	"io"
)

const gzipEncoding = "gzip"

func init() {
	defaultCompressors = append(
		defaultCompressors,
		&GzipCompressor{gzip.DefaultCompression},
	)
}

type GzipCompressor struct {
	Level int
}

func (c *GzipCompressor) Encoding() string {
	return gzipEncoding
}

func (c *GzipCompressor) Weight() int {
	return 200
}

func (c *GzipCompressor) NewWriter(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, c.Level)
}

func WithGzip(level int) Configuration {
	return func(c *Config) {
		c.deleteEncoding(gzipEncoding)
		c.addCompressor(&GzipCompressor{level})
	}
}
