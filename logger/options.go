package logger

type Options struct {
	version string
	pretty  bool
}

type OptionFn func(options *Options)

func WithVersion(version string) OptionFn {
	return func(options *Options) {
		options.version = version
	}
}

func WithPretty() OptionFn {
	return func(options *Options) {
		options.pretty = true
	}
}
