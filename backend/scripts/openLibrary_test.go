package scripts_test

import (
	"bookTracker/scripts"
	"context"
	"testing"
)

func BenchmarkParseAuthor(b *testing.B) {
	line := `OL1A    \t \t \t \t {"key":"OL1A","name":"Author Name"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scripts.ParseAuthor(line)
	}
}

func BenchmarkScanFile(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		scripts.ImportFile(ctx, nil, "authors_sample.txt", scripts.ParseAuthor)
	}
}
