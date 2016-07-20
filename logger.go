package laozi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const maxRetries = 10

// Logger defines the behaviour of all loggers.
type Logger interface {
	// send event data
	Log([]byte)
	// Close method called before it removed from internal map
	Close() error
	// LastActive is used to get the time a logger last logged. Used for deleting stale
	// loggers from internal map.
	LastActive() time.Time
}

type s3logger struct {
	S3            *s3.S3
	bucket        string
	key           string
	buffer        *bytes.Buffer
	active        time.Time
	logChan       chan []byte
	flushInterval time.Duration
	quitChan      chan struct{}
	compression   string
}

// Log causes event event to br written to internal memory buffer.
func (l *s3logger) Log(e []byte) {
	l.logChan <- e
	l.active = time.Now()
}

func (l *s3logger) loop() {
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
			l.buffer.Write(event)
		case <-l.quitChan:
			return
		default:
			// chill out for a moment...
			time.Sleep(time.Millisecond)
		}
	}
}

// Close is called when logger timeouts. Will cause internal memory buffer to be written to s3.
func (l *s3logger) Close() error {
	l.quitChan <- struct{}{}
	return l.flush()
}

func (l *s3logger) compressBuffer() (bs []byte) {

	switch l.compression {
	case "gzip":
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		w.Write(l.buffer.Bytes())
		w.Close()
		bs = b.Bytes()
	case "":
		bs = l.buffer.Bytes()
	}
	return
}

func (l *s3logger) decompressToBuffer(r io.ReadCloser) {

	switch l.compression {
	case "gzip":
		gr, _ := gzip.NewReader(r)
		b, _ := ioutil.ReadAll(gr)
		l.buffer.Write(b)
	case "":
		b, _ := ioutil.ReadAll(r)
		l.buffer.Write(b)
	}
	return
}

func (l *s3logger) flush() error {

	var err error
	// retry write to s3 for max tries
	for i := 0; i < maxRetries; i++ {
		_, err = l.S3.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(l.bucket),
			Key:    aws.String(l.key),
			Body:   bytes.NewReader(l.compressBuffer()),
		})

		if err == nil {
			break
		}
	}

	// TODO: add emergency file writing here if s3 is down...
	// if err != nil {
	// 	return err
	// }

	return err
}

// LastActive is used to know when the S3Logger last logged.
func (l *s3logger) LastActive() time.Time {
	return l.active
}

// fetchPreviousData will go fetch any previous data stored on s3 for a corresponding key
func (l *s3logger) fetchPreviousData() {
	resp, err := l.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.key),
	})

	if err != nil {
		// TODO: what to do with error
		fmt.Println(err)
	}

	if resp.Body != nil {
		l.decompressToBuffer(resp.Body)
	}
}
