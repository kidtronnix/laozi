package main

import (
	"fmt"
	"time"

	laozi "github.com/seedboxtech/laozi"
)

func main() {
	l := laozi.NewLaozi(&laozi.Config{
		LoggerFactory: laozi.S3LoggerFactory{
			Bucket:        "laozi-test",
			Prefix:        "events/", // optional
			Region:        "us-east-1",
			FlushInterval: time.Second * 30,
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

	// give some time for the final flush to happen
	<-time.After(2 * time.Minute)

}