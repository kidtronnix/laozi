## laozi

[![GoDoc](https://godoc.org/github.com/seedboxtech/laozi?status.svg)](https://godoc.org/github.com/seedboxtech/laozi)

archiver of events.

stores events to s3 partitioned however you want. imagine AWS firehose service but with a
configurable partition method.

## usage

```go
package main

import "github.com/seedboxtech/laozi"

func main() {
  // create a new laozi instance
  l := laozi.NewLaozi(&loazi.Config{
    // LoggerFactory is a factory that will generate new loggers as they are needed.
    // each new logger is passed the partition key it is responsible for logging data for.
    LoggerFactory: loazi.S3LoggerFactory{
      Bucket: "my-s3-bucket",
      Prefix: "events/", // optional
      Region: "us-east-1",
    },
    // EventChanSize is the buffer size of the event channel
    EventChanSize: 10000,
     // LoggerTimeout is the timeout to wait for logger to be inactive before it's `Close` method is called
    LoggerTimeout: time.Minute,
    // PartitionKeyFunc is a func you must implement to extract a key a logged data point
    PartitionKeyFunc: func(e []byte) (string, error) {
      // in this example we use the first value in a comma seperated list of values as the partition key
      parts := bytes.Split(e, []byte(","))
      return fmt.Sprintf("%s", parts[0]), nil
    },
  })

  // tell loazi about your event, function is totally non-blocking internally
  // so no waiting around for event to be processed.
  l.Log([]byte("key,some data to store line by line\n"))
}

```

## testing

currently this package uses s3 directly in tests. this does mean tests will cost a very small
amount of money to run and requires a connection to the internet. tests can be run...

```bash
go test ./...
```
