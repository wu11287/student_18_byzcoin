package pow

import (
	"fmt"
	"testing"
)

func BenchmarkPow(b *testing.B) {
	for i := 0; i < b.N; i++{
		var end float64
		for j := 0; j < 3; j++ {
			end += powMain()
		}
		fmt.Println(end/3.0)
		// end := powMain()
		// fmt.Println(end)
	}
}