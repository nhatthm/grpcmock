package must

// NotFail throws a panic if there is an error.
func NotFail(err error) {
	if err != nil {
		panic(err)
	}
}
