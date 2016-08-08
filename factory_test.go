package laozi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerFactoryNew(t *testing.T) {
	assert := assert.New(t)

	lf := S3LoggerFactory{
		Bucket:        "bucket",
		Prefix:        "prefix/",
		Region:        "us-east-1",
		FlushInterval: 100,
	}

	l := lf.NewLogger("test.file")

	assert.Implements((*Logger)(nil), l)
}

func TestLoggerFactoryNewDedupedLogger(t *testing.T) {
	assert := assert.New(t)

	lf := S3LoggerFactory{
		Bucket:        "bucket",
		Prefix:        "prefix/",
		Region:        "us-east-1",
		FlushInterval: 100,
		IsDupeFunc:    func(event []byte, line []byte) bool { return string(event) == string(line) },
	}

	l := lf.NewLogger("test.file")

	assert.Implements((*Logger)(nil), l)
}
