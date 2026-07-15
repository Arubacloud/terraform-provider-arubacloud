package aruba

func NewClient(options *Options) (Client, error) {
	return buildClient(options)
}

// TODO: Two variations of `NewClient`
// - `NewClient()`: returns a client with the default config
// - `NewClientWithOptions(opts Options)`: returns a client with a custom config

// TODO: `DefaultOptions()` function to return a `Options` instance with default values

// TODO: `LoadOptionsFromPath(path path.Path)` function to allow to read option values from a file

// TODO: `LoadOptionsFromURL(url net.URL)` function to allow to read option values from a web server
