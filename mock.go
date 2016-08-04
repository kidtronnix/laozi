package laozi

import "fmt"

type MockLaozi struct{}

func (d MockLaozi) Log(b []byte) {
	fmt.Printf("[laozi] event logged: %s\n", b)
}

func (d MockLaozi) Close() {
	fmt.Println("[laozi] closing!")
}
