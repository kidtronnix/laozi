package laozi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDedupeLoopLogsUniqueEvents(t *testing.T) {
	assert := assert.New(t)

	l := makeTestLogger()

	dl := dedupeS3Logger{l, func(event []byte, line []byte) bool { return string(event) == string(line) }}

	go dl.loop()

	dl.logChan <- []byte("a\n")
	dl.logChan <- []byte("b\n")
	dl.logChan <- []byte("b\n")
	dl.logChan <- []byte("c\n")
	dl.logChan <- []byte("a\n")
	dl.logChan <- []byte("c\n")
	dl.logChan <- []byte("b\n")
	dl.logChan <- []byte("c\n")
	dl.logChan <- []byte("c\n")

	time.Sleep(time.Millisecond * 15)

	assert.Equal("a\nb\nc\n", string(dl.buffer.Bytes()))
}
