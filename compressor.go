package compression

import "io"

var defaultCompressors Compressors

type Compressor interface {
	Encoding() string
	Weight() int
	NewWriter(w io.Writer) (io.WriteCloser, error)
}

type baseCompressor struct {
	encoding string
	weight   int
}

func (c *baseCompressor) Encoding() string {
	return c.encoding
}
func (c *baseCompressor) Weight() int {
	return c.weight
}

type Compressors []Compressor

// Len is the number of elements in the collection.
func (c Compressors) Len() int {
	return len(c)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (c Compressors) Less(i, j int) bool {
	return c[i].Weight() < c[j].Weight()
}

// Swap swaps the elements with indexes i and j.
func (c Compressors) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Compressors) Find(encoding string) Compressor {
	for _, item := range c {
		if item.Encoding() == encoding {
			return item
		}
	}
	return nil
}
