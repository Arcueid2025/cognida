package retriever

import "testing"

func TestChineseAnalyzerParams(t *testing.T) {
	if got := chineseAnalyzerParams["type"]; got != "chinese" {
		t.Fatalf("BM25 text analyzer type = %v, want chinese", got)
	}
	if len(chineseAnalyzerParams) != 1 {
		t.Fatalf("chinese analyzer must not carry unsupported optional params: %#v", chineseAnalyzerParams)
	}
}
