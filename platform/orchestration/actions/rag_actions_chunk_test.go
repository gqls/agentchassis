package actions

import (
	"strings"
	"testing"
)

// chunkContent previously looped forever on any content longer than chunkSize:
// the final chunk ends at len(content), start = end - overlap re-enters the
// loop, and the same tail is appended until OOM (chassis OOMKills of
// 2026-07-09/10, both inside rag_index on a ~3KB PLAN body).
func TestChunkContentTerminatesOnLongContent(t *testing.T) {
	// Sentence-ish content just over the size that triggered the incident.
	content := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 70) // ~3,200 chars
	chunks := chunkContent(content, 1000, 200)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	// ~3.2KB at ~800 net chars per chunk: a handful, never hundreds.
	if len(chunks) > 10 {
		t.Fatalf("suspicious chunk count %d — tail-loop regression?", len(chunks))
	}
	last := chunks[len(chunks)-1]
	if !strings.HasSuffix(strings.TrimSpace(content), last) {
		t.Errorf("last chunk should be the content tail, got %q", last[:min(40, len(last))])
	}
	for i, c := range chunks {
		if c == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

func TestChunkContentShortContentSingleChunk(t *testing.T) {
	chunks := chunkContent("short content", 1000, 200)
	if len(chunks) != 1 || chunks[0] != "short content" {
		t.Fatalf("expected the content back as one chunk, got %v", chunks)
	}
}

func TestChunkContentNoSentenceBoundaries(t *testing.T) {
	content := strings.Repeat("x", 2500) // no '.' or '\n' anywhere
	chunks := chunkContent(content, 1000, 200)
	if len(chunks) == 0 || len(chunks) > 5 {
		t.Fatalf("expected a few chunks, got %d", len(chunks))
	}
}

func TestChunkContentPathologicalOverlap(t *testing.T) {
	// overlap == chunkSize used to be another non-terminating shape; the
	// forward-progress guard must handle any config the workflow passes.
	content := strings.Repeat("y", 3000)
	chunks := chunkContent(content, 1000, 1000)
	if len(chunks) == 0 || len(chunks) > 10 {
		t.Fatalf("expected bounded chunks under pathological overlap, got %d", len(chunks))
	}
}
