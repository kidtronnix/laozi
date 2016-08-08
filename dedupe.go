package laozi

import (
	"io"
	"time"
)

type dedupeS3Logger struct {
	*s3logger
	isDupeFunc func(event []byte, line []byte) bool
}

func (l *dedupeS3Logger) loop() {
	var flushChan <-chan time.Time
	if l.flushInterval > 0 {
		flushChan = time.After(l.flushInterval)
	}

	var event []byte
	for {
		select {
		case <-flushChan:
			l.flush()
			if l.flushInterval > 0 {
				flushChan = time.After(l.flushInterval)
			}
		case event = <-l.logChan:
			var tmp []byte
			for {
				line, err := l.buffer.ReadBytes('\n')
				if err == io.EOF {
					// didn't find dupe in buffer so write
					tmp = append(tmp, event...)
					break
				}
				if l.isDupeFunc(event, line) {
					tmp = append(append(tmp, line...), l.buffer.Bytes()...)
					break
				}
				tmp = append(tmp, line...)
			}
			l.buffer.Reset()
			l.buffer.Write(tmp)
		case <-l.quitChan:
			return
		default:
			// chill out for a moment...
			time.Sleep(time.Millisecond)
		}
	}
}
