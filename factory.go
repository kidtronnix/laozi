package laozi

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// LoggerFactory is an interface that defines how to make a new logger.
// This Logger will be responsible for logging all events to it that match the same
// partition key.
type LoggerFactory interface {
	NewLogger(key string) Logger
}

// S3LoggerFactory is a logger factory for creating loggers that log received events to S3.
type S3LoggerFactory struct {
	Prefix        string
	Bucket        string
	Region        string
	FlushInterval time.Duration
}

// NewLogger return a new instance of an S3 Logger for a corresponding partition key.
func (lf S3LoggerFactory) NewLogger(key string) Logger {
	l := &s3logger{
		bucket:      lf.Bucket,
		key:         fmt.Sprintf("%s%s", lf.Prefix, key),
		S3:          s3.New(session.New(), &aws.Config{Region: aws.String(lf.Region)}),
		buffer:      bytes.NewBuffer([]byte{}),
		active:      time.Now(),
		logChan:     make(chan []byte),
		quitChan:    make(chan struct{}),
		flushTicker: time.NewTicker(lf.FlushInterval),
	}
	l.fetchPreviousData()
	go l.loop()

	return l

}
