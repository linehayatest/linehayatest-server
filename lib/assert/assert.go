package assert

import (
	"github.com/lainio/err2/assert"
)

func NoErrorf(err error, format string, args ...interface{}) {
	format = format + ": %v"
	args = append(args, err)
	assert.P.Truef(err == nil, format, args...)
}

func NoError(err error, prefix string) {
	if prefix != "" {
		prefix = prefix + ": %v"
	}
	assert.P.Truef(err == nil, prefix, err)
}
