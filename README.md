## laozi

[![GoDoc](https://godoc.org/github.com/seedboxtech/laozi?status.svg)](https://godoc.org/github.com/seedboxtech/laozi)

archiver of events.

stores events to s3 partitioned however you want. imagine AWS firehose service but with a
configurable partition method.

## usage

```go
package main

import (
	"fmt"
	"time"

	laozi "github.com/seedboxtech/laozi"
)

func main() {
	l := laozi.NewLaozi(&laozi.Config{
		LoggerFactory: laozi.S3LoggerFactory{
			Bucket: "laozi-test",
			Region: "us-east-1",
			Prefix: "events/", // optional
			FlushInterval: time.Second * 30, // optional
			Compression: "gzip", // optional
		},
		EventChannelSize: 10000000,
		LoggerTimeout:    time.Minute,
		PartitionKeyFunc: func(e []byte) (string, error) {
			return "event-file.csv.gz", nil
		},
	})

	quit := time.After(2 * time.Minute)
	i := 0
loop:
	for {
		select {
		case <-quit:
			break loop
		default:
			i++
			l.Log([]byte(fmt.Sprintf("%d\n", i)))
			time.Sleep(time.Millisecond)

		}
	}

	fmt.Println("Done logging!")
	fmt.Println(" - logged:", i)

	// give some time for the final flush to happenx
	<-time.After(2 * time.Minute)

}


```

## testing

currently this package uses s3 directly in tests. this does mean tests will cost a very small
amount of money to run and requires a connection to the internet. tests can be run...

```bash
go test ./...
```
