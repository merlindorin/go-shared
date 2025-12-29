package must

// Get returns v as is. It panics if err is non-nil.
// example: must.Get(url.Parse("something"))
func Get[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
