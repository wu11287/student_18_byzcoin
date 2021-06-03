package test

import (
	"fmt"
	"testing"
)

func BenchmarkVerfied(b *testing.B) {
	var cnt int = 0
	for i := 0; i < b.N; i++ {
		ok := verified()
		if ok {
			cnt++
		}
	}
	fmt.Println(cnt)
}