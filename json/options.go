package json

type decodeOptions struct {
	maxBodyBytes int64
	strict       bool
}

// Option задаёт поведение DecodeJSON.
type Option func(*decodeOptions)

// DefaultMaxBodyBytes — дефолтный лимит размера тела запроса.
const DefaultMaxBodyBytes int64 = 1 << 20

// WithMaxBodyBytes задаёт максимальный размер тела запроса.
func WithMaxBodyBytes(n int64) Option {
	return func(opts *decodeOptions) {
		if n > 0 {
			opts.maxBodyBytes = n
		}
	}
}
