package testmi

import (
	"fmt"
	"testing"
)

func BenchmarkMain(b *testing.B) {
	var cnt int = 0
	for i:=0;i<b.N;i++{
		ok := count()
		if !ok {
			cnt++
		}
	}
	fmt.Println(cnt)
}