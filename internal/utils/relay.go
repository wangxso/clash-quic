// 数据转发工具
package utils

import (
	"io"
	"sync"
)

// 在两个读写器之间转发数据
func Relay(a, b io.ReadWriter) {
	var wg sync.WaitGroup
	wg.Add(2)

	// a -> b
	go func() {
		defer wg.Done()
		io.Copy(b, a)
		if closer, ok := b.(io.Closer); ok {
			closer.Close()
		}
	}()

	// b -> a
	go func() {
		defer wg.Done()
		io.Copy(a, b)
		if closer, ok := a.(io.Closer); ok {
			closer.Close()
		}
	}()

	wg.Wait()
}
