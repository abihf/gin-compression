package compression

import "sort"

type Config struct {
	MinSize     int
	Compressors Compressors

	supportedEncodings []string
}

func (c *Config) addCompressor(comp Compressor) {
	c.Compressors = append(c.Compressors, comp)
}

func (c *Config) deleteEncoding(encoding string) {
	newList := Compressors{}
	for _, compressor := range c.Compressors {
		excluded := false

		if compressor.Encoding() == encoding {
			excluded = true
		}
		if !excluded {
			newList = append(newList, compressor)
		}
	}
	c.Compressors = newList
}

type Configuration func(c *Config)

func initConfig(configFns []Configuration) *Config {
	config := &Config{
		MinSize:     16384, // 16k
		Compressors: Compressors{},
	}
	config.Compressors = defaultCompressors[:]

	for _, fn := range configFns {
		fn(config)
	}

	sort.Sort(config.Compressors)
	config.supportedEncodings = []string{}
	for _, compressor := range config.Compressors {
		config.supportedEncodings = append(config.supportedEncodings, compressor.Encoding())
	}

	return config
}

func WithMinSize(minSize int) Configuration {
	return func(c *Config) {
		c.MinSize = minSize
	}
}

// WithCompressor add custom compressor
func WithCompressor(compressor Compressor) Configuration {
	return func(c *Config) {
		c.addCompressor(compressor)
	}
}

// WithoutEncodings remove already registered encodings
func WithoutEncodings(encodings ...string) Configuration {
	return func(c *Config) {
		newList := Compressors{}
		for _, compressor := range c.Compressors {
			excluded := false
			for _, encoding := range encodings {
				if compressor.Encoding() == encoding {
					excluded = true
				}
			}
			if !excluded {
				newList = append(newList, compressor)
			}
		}
		c.Compressors = newList
	}
}
