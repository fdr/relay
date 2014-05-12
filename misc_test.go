package relay_test

import (
	"bytes"
)

type nopReadWriteCloser struct {
	bytes.Buffer
}

func (c *nopReadWriteCloser) Close() error {
	return nil
}
