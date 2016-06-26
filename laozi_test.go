package laozi

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testTime = time.Now()

func MockPartitionFunc(e []byte) (string, error) {
	// just uses event data as partition key
	return fmt.Sprintf("%s", e), nil
}

type MockLoggerFactory struct{}

func (mf *MockLoggerFactory) NewLogger(file string) Logger {
	ml := &MockLogger{
		fileName: file,
		bytes:    make([]byte, 0),
	}
	return ml
}

type MockLogger struct {
	fileName string
	bytes    []byte
	closed   bool
}

type MockLoggerCloseError struct {
	MockLogger
}

func (m *MockLogger) Log(b []byte) {
	m.bytes = append(m.bytes, b...)
}

func (m MockLogger) LastActive() time.Time {
	return testTime
}

func (m *MockLogger) Close() error {
	m.closed = true
	return nil
}

func (m *MockLoggerCloseError) Close() error {
	return errors.New("Couldnt close logger!")
}

func TestLaoziLog(t *testing.T) {
	assert := assert.New(t)
	l := &laozi{
		EventChan: make(chan []byte, 1),
	}
	e := []byte("1")
	l.Log(e)

	assert.Equal(e, <-l.EventChan)
}

func TestRouterSkipsOnBadPartition(t *testing.T) {
	assert := assert.New(t)

	l := &laozi{
		EventChan:  make(chan []byte),
		routingMap: map[string]Logger{},
		Config: &Config{
			LoggerFactory:    &MockLoggerFactory{},
			LoggerTimeout:    time.Minute,
			PartitionKeyFunc: func([]byte) (string, error) { return "", errors.New("Could not generate partition key!") },
		},
	}
	go l.route()

	l.EventChan <- []byte("1")
	l.EventChan <- []byte("2")
	l.EventChan <- []byte("3")

	assert.Equal(0, len(l.routingMap))
}

func TestRouterCreatesLoggers(t *testing.T) {
	assert := assert.New(t)

	l := &laozi{
		EventChan:  make(chan []byte),
		routingMap: map[string]Logger{},
		Config: &Config{
			LoggerFactory:    &MockLoggerFactory{},
			LoggerTimeout:    time.Minute,
			PartitionKeyFunc: MockPartitionFunc,
		},
	}
	go l.route()

	l.EventChan <- []byte("1")
	l.EventChan <- []byte("2")
	l.EventChan <- []byte("1")

	assert.Equal(2, len(l.routingMap))
}

func TestRouterCallsLog(t *testing.T) {
	assert := assert.New(t)

	l := &laozi{
		EventChan:  make(chan []byte),
		routingMap: map[string]Logger{},
		Config: &Config{
			LoggerFactory:    &MockLoggerFactory{},
			LoggerTimeout:    10 * time.Millisecond,
			PartitionKeyFunc: MockPartitionFunc,
		},
	}
	go l.route()

	l.EventChan <- []byte("1")
	l.EventChan <- []byte("1")

	logger := l.routingMap["1"].(*MockLogger)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(logger.bytes, []byte("11"))
}

func TestRouterDeletesLoggersAfterTimeout(t *testing.T) {
	assert := assert.New(t)
	l := &laozi{
		routingMap: map[string]Logger{},
		Config: &Config{
			LoggerTimeout: 2 * time.Millisecond,
		},
	}

	log1 := &MockLogger{}
	log2 := &MockLogger{}
	l.routingMap["testkey1"] = log1
	l.routingMap["testkey2"] = log2

	go l.monitorLoggers()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(0, len(l.routingMap))
	assert.True(log1.closed)
	assert.True(log2.closed)
}

func TestRouterCloses(t *testing.T) {
	assert := assert.New(t)

	l := &laozi{
		routingMap: map[string]Logger{},
	}

	log1 := &MockLogger{}
	log2 := &MockLogger{}
	l.routingMap["testkey1"] = log1
	l.routingMap["testkey2"] = log2

	l.Close()
	assert.True(log1.closed)
	assert.True(log2.closed)
}

func TestRouterClosesError(t *testing.T) {
	// assert := assert.New(t)

	l := &laozi{
		routingMap: map[string]Logger{},
	}

	log1 := &MockLoggerCloseError{MockLogger{}}
	l.routingMap["testkey1"] = log1

	l.Close()
}

func TestNewLoazi(t *testing.T) {
	assert := assert.New(t)

	l := NewLaozi(&Config{
		LoggerTimeout: time.Minute,
		LoggerFactory: S3LoggerFactory{
			Bucket: "bucket",
			Prefix: "prefix/",
			Region: "us-east-1",
		},
		PartitionKeyFunc: func([]byte) (string, error) {
			return "a", nil
		},
	})

	assert.Implements((*Laozi)(nil), l)
}
