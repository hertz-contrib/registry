package test_registry

import "errors"

func Add(a, b int) int {
	E()
	return a + b
}

func E() error {
	return errors.New("1111")
}
