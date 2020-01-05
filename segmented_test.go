package cache

import "testing"

func BenchmarkSegmented(b *testing.B) {
	s := NewSegmented(NewFIFO(10), NewLRU(10))
	b.ResetTimer()

	const c = 10000

	for i := 0; i < b.N; i++ {
		s.Add(i%(c*3), 3)
	}

	b.StopTimer()
	_ = s.Close()
}
