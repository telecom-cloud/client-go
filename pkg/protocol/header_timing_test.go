package protocol

import (
	"net/http"
	"strconv"
	"testing"
)

func BenchmarkHTTPHeaderGet(b *testing.B) {
	hh := make(http.Header)
	hh.Set("X-tt-logid", "abc123456789")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hh.Get("X-tt-logid")
	}
}

func BenchmarkCrafterHeaderGet(b *testing.B) {
	zh := new(ResponseHeader)
	zh.Set("X-tt-logid", "abc123456789")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zh.Get("X-tt-logid")
	}
}

func BenchmarkHTTPHeaderSet(b *testing.B) {
	hh := make(http.Header)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hh.Set("X-tt-logid", "abc123456789")
	}
}

func BenchmarkCrafterHeaderSet(b *testing.B) {
	zh := new(ResponseHeader)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zh.Set("X-tt-logid", "abc123456789")
	}
}

func BenchmarkHTTPHeaderAdd(b *testing.B) {
	hh := make(http.Header)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hh.Add("X-tt-"+strconv.Itoa(i), "abc123456789")
	}
}

func BenchmarkCrafterHeaderAdd(b *testing.B) {
	zh := new(ResponseHeader)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zh.Add("X-tt-"+strconv.Itoa(i), "abc123456789")
	}
}

func BenchmarkRefreshServerDate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		refreshServerDate()
	}
}
