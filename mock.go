package laozi

import "fmt"

type MockLaozi struct{}

func (d MockLaozi) Log(b []byte) {
	fmt.Println("Mock Laozi output", b)
}

func (d MockLaozi) Close() {
	fmt.Println("Mock Laozi closing")
}
