// +build cgo

package compression

import (
	"io"

	brotli "github.com/google/brotli/go/cbrotli"
)

const brotliEncoding = "br"

func init() {
	defaultCompressors = append(
		defaultCompressors,
		&BrotliCompressor{brotli.WriterOptions{
			Quality: 4,
			LGWin:   19,
		}},
	)
}

type BrotliCompressor struct {
	Options brotli.WriterOptions
}

func (c *BrotliCompressor) Encoding() string {
	return brotliEncoding
}

func (c *BrotliCompressor) Weight() int {
	return 100
}

func (c *BrotliCompressor) NewWriter(w io.Writer) (io.WriteCloser, error) {
	return brotli.NewWriter(w, c.Options), nil
}

func WithBrotli(options brotli.WriterOptions) Configuration {
	return func(c *Config) {
		c.deleteEncoding(brotliEncoding)
		c.addCompressor(&BrotliCompressor{options})
	}
}
