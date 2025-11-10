package federation

import (
	"fmt"
	"testing"
)

// BenchmarkDetectRelationships benchmarks relationship detection with varying sizes
func BenchmarkDetectRelationships(b *testing.B) {
	detector := NewDetector()

	tests := []struct {
		name        string
		resultCount int
	}{
		{"100_results", 100},
		{"500_results", 500},
		{"1000_results", 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			results := generateTestQueryResults(tt.resultCount)
			queryItems := generateTestItems(tt.resultCount)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detector.DetectRelationships(results, queryItems)
			}
		})
	}
}

// BenchmarkBuildRelationshipGraph benchmarks graph construction
func BenchmarkBuildRelationshipGraph(b *testing.B) {
	detector := NewDetector()
	relationships := map[string][]string{
		"item-1": {"item-2", "item-3", "item-4"},
		"item-2": {"item-1", "item-5"},
		"item-3": {"item-1", "item-6", "item-7"},
		"item-4": {"item-1"},
		"item-5": {"item-2"},
		"item-6": {"item-3"},
		"item-7": {"item-3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.BuildRelationshipGraph(relationships)
	}
}

// BenchmarkMergerAddResults benchmarks adding results to merger
func BenchmarkMergerAddResults(b *testing.B) {
	tests := []struct {
		name        string
		resultCount int
	}{
		{"100_results", 100},
		{"500_results", 500},
		{"1000_results", 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			items := generateTestItems(tt.resultCount)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				merger := NewMerger()
				merger.AddResults("source-1", items)
				merger.AddResults("source-2", items)
				merger.AddResults("source-3", items)
			}
		})
	}
}

// BenchmarkMergerMergeAndDeduplicate benchmarks deduplication
func BenchmarkMergerMergeAndDeduplicate(b *testing.B) {
	tests := []struct {
		name        string
		resultCount int
	}{
		{"100_results", 100},
		{"500_results", 500},
		{"1000_results", 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			items := generateTestItems(tt.resultCount)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				merger := NewMerger()
				merger.AddResults("source-1", items)
				merger.AddResults("source-2", items)
				merger.AddResults("source-3", items)
				_, _ = merger.MergeAndDeduplicate()
			}
		})
	}
}

// BenchmarkMergerHashContent benchmarks content hashing
func BenchmarkMergerHashContent(b *testing.B) {
	merger := NewMerger()
	content := "This is a sample document content for hashing purposes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.hashContent(content)
	}
}

// BenchmarkMergerItemToString benchmarks item-to-string conversion
func BenchmarkMergerItemToString(b *testing.B) {
	merger := NewMerger()
	item := map[string]interface{}{
		"id":    "item-123",
		"title": "Sample Document",
		"text":  "This is the content",
		"score": 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.itemToString(item)
	}
}

// BenchmarkDetectorIsRelated benchmarks relationship detection
func BenchmarkDetectorIsRelated(b *testing.B) {
	detector := NewDetector()
	item1 := map[string]interface{}{
		"id":    "item-1",
		"title": "Kubernetes deployment",
	}
	item2 := map[string]interface{}{
		"id":    "item-2",
		"title": "Kubernetes in production",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.isRelated(item1, item2)
	}
}

// BenchmarkDetectorHasSimilarContent benchmarks content similarity
func BenchmarkDetectorHasSimilarContent(b *testing.B) {
	detector := NewDetector()
	str1 := "Kubernetes deployment guide for production environments"
	str2 := "Kubernetes in production deployment guide"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.hasSimilarContent(str1, str2)
	}
}

// Helper functions

func generateTestQueryResults(count int) []*QueryResult {
	sources := []string{"source-a", "source-b", "source-c"}
	results := make([]*QueryResult, len(sources))

	itemsPerSource := count / len(sources)
	for s, source := range sources {
		items := generateTestItems(itemsPerSource)
		results[s] = &QueryResult{
			Source:    source,
			Items:     items,
			ItemCount: len(items),
		}
	}

	return results
}

func generateTestItems(count int) []interface{} {
	items := make([]interface{}, count)
	titles := []string{
		"Kubernetes deployment guide",
		"Docker container orchestration",
		"Microservices architecture",
		"Cloud infrastructure automation",
		"CI/CD pipeline design",
		"Application monitoring",
		"Performance optimization",
		"Security best practices",
	}

	contents := []string{
		"Learn how to deploy and manage containerized applications",
		"Best practices for cloud infrastructure",
		"Guide to building scalable systems",
		"Production deployment strategies",
		"DevOps automation and tooling",
		"Enterprise architecture patterns",
		"Reliability and high availability",
		"Distributed systems design",
	}

	for i := 0; i < count; i++ {
		items[i] = map[string]interface{}{
			"id":       fmt.Sprintf("item-%d", i),
			"title":    titles[i%len(titles)],
			"content":  contents[i%len(contents)],
			"source":   fmt.Sprintf("source-%c", 65+(i%3)), // source-A, source-B, or source-C
			"metadata": map[string]interface{}{"index": i, "score": 0.95},
		}
	}
	return items
}
