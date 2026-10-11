package benchmarks

import (
	"bytes"
	"io"
	"testing"
)

const benchmarkPayloadSize = 4 << 20

func BenchmarkStreamCopy(b *testing.B) {
	payload := bytes.Repeat([]byte("gfw404-ultra"), benchmarkPayloadSize/len("gfw404-ultra"))
	buf := make([]byte, 32<<10)
	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := io.CopyBuffer(io.Discard, bytes.NewReader(payload), buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDatagramFragmentation(b *testing.B) {
	payload := bytes.Repeat([]byte("gfw404-ultra"), (32<<10)/len("gfw404-ultra"))
	const maxPayload = 1200
	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragments := fragmentPayload(payload, maxPayload)
		if len(fragments) == 0 {
			b.Fatal("fragmentation produced no fragments")
		}
	}
}

func fragmentPayload(payload []byte, maxPayload int) [][]byte {
	fragments := make([][]byte, 0, (len(payload)+maxPayload-1)/maxPayload)
	for offset := 0; offset < len(payload); offset += maxPayload {
		end := offset + maxPayload
		if end > len(payload) {
			end = len(payload)
		}
		fragment := make([]byte, end-offset)
		copy(fragment, payload[offset:end])
		fragments = append(fragments, fragment)
	}
	return fragments
}
