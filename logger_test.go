package laozi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var testBucket = fmt.Sprintf("LOAZI_TEST_BUCKET_%d", rand.Intn(123435))
var testFile = "TEST_LOGGER_FILE.gz"

func makeS3Service() *s3.S3 {
	return s3.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
}

func makeTestLogger() *s3logger {
	return &s3logger{
		bucket:    testBucket,
		key:       testFile,
		S3:        makeS3Service(),
		buffer:    bytes.NewBuffer([]byte{}),
		active:    time.Now(),
		logChan:   make(chan []byte, 1),
		quitChan:  make(chan struct{}, 1),
		flushChan: make(<-chan time.Time),
	}
}

func makeTestBucket() {

	svc := makeS3Service()
	// create test bucket
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: &testBucket,
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func detroyTestBucket() {

	svc := makeS3Service()

	svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &testBucket,
		Key:    &testFile,
	})
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &testBucket,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

}

func TestS3LoggerLog(t *testing.T) {
	assert := assert.New(t)

	l := makeTestLogger()
	testData := []byte("some data")
	l.Log(testData)

	assert.Equal(testData, <-l.logChan)

	assert.WithinDuration(time.Now(), l.active, time.Millisecond)
}

func TestS3LoggerLastActive(t *testing.T) {
	assert := assert.New(t)

	now := time.Now()
	l := makeTestLogger()
	l.active = now

	assert.Equal(l.LastActive(), now)
}

func TestS3LoggerCloses(t *testing.T) {
	assert := assert.New(t)

	testData := []byte("some data")
	makeTestBucket()
	defer detroyTestBucket()

	l := makeTestLogger()
	l.buffer.Write(testData)
	err := l.Close()

	assert.NoError(err)

	// do a read test to check
	resp, err := l.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.key),
	})
	assert.NoError(err)

	bs, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(testData)
	w.Close()

	assert.Equal(b.Bytes(), bs)

}

func TestS3LoggerFetchesPreviousData(t *testing.T) {
	assert := assert.New(t)

	testData := []byte("some data")
	makeTestBucket()
	defer detroyTestBucket()

	l := makeTestLogger()

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()
	w.Write(testData)
	w.Close()

	// put existing file on s3 so we can check it fetches data from here
	_, err := l.S3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.key),
		Body:   bytes.NewReader(b.Bytes()),
	})
	assert.NoError(err)

	l.fetchPreviousData()

	assert.Equal(testData, l.buffer.Bytes())

}

func TestLoopLogsEvent(t *testing.T) {
	assert := assert.New(t)

	testData := []byte("test data")
	l := makeTestLogger()

	go l.loop()

	l.logChan <- testData

	time.Sleep(time.Millisecond * 5)

	assert.Equal(testData, l.buffer.Bytes())

}

func TestLoopFlushes(t *testing.T) {
	assert := assert.New(t)

	makeTestBucket()
	defer detroyTestBucket()

	testData := []byte("test data")
	l := makeTestLogger()
	l.buffer.Write(testData)
	go l.loop()
	l.flushChan = time.After(time.Millisecond)

	// this needs to be long enough for file to write to s3
	time.Sleep(time.Second * 5)

	// do a read test to check
	resp, err := l.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.key),
	})
	assert.NoError(err)

	bs, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(testData)
	w.Close()

	assert.Equal(b.Bytes(), bs)

}
