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
