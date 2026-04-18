package gozod

// Apply integrates external functions into schema chains.
func Apply[S any, R any](schema S, fn func(S) R) R {
	return fn(schema)
}
