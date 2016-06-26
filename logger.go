package laozi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const MaxRetries = 10

type Logger interface {
	Log([]byte)
	Close() error
	LastActive() time.Time
}

type s3logger struct {
	S3     *s3.S3
	bucket string
	key    string
	buffer *bytes.Buffer
	active time.Time
}

// Log adds receives data for archiving
func (l *s3logger) Log(b []byte) {
	l.buffer.Write(b)
	l.active = time.Now()
}

// will cause bytes to be written to s3
func (l *s3logger) Close() error {

	// b is a buffer where we put bytes to and read bytes from
	var b bytes.Buffer
	// make a gzip writer that can write to buffer
	w := gzip.NewWriter(&b)
	// dump contents of loggers
	w.Write(l.buffer.Bytes())
	// gzip needs to be closed as it's compression can only run at the end
	w.Close()

	var err error
	// retry write to s3 for max tries
	for i := 0; i < MaxRetries; i++ {
		_, err = l.S3.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(l.bucket),
			Key:    aws.String(l.key),
			Body:   bytes.NewReader(b.Bytes()),
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

func (l *s3logger) LastActive() time.Time {
	return l.active
}

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
		gr, _ := gzip.NewReader(resp.Body)
		b, _ := ioutil.ReadAll(gr)
		l.buffer.Write(b)
	}
}
