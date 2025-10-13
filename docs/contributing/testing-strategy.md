# Testing Strategy

**Conexus (Agentic Context Engine)**  
**Version:** 1.0  
**Last Updated:** October 12, 2025

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Testing Pyramid](#testing-pyramid)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [Performance Testing](#performance-testing)
- [RAG Evaluation Testing](#rag-evaluation-testing)
- [Test Data Management](#test-data-management)
- [Coverage Requirements](#coverage-requirements)
- [CI/CD Integration](#cicd-integration)
- [Best Practices](#best-practices)

---

## Testing Philosophy

### Core Principles

1. **Test-Driven Development (TDD)** - Write tests before implementation
2. **Fast Feedback** - Tests should run quickly and fail fast
3. **Isolation** - Tests should not depend on external state or each other
4. **Reliability** - Tests should be deterministic and reproducible
5. **Maintainability** - Tests should be as readable as production code

### Testing Goals

- **80-90% Code Coverage** - Comprehensive test coverage for all critical paths
- **<5 minute test suite** - Fast feedback for developers
- **Zero flaky tests** - Reliable CI/CD pipeline
- **Quality over quantity** - Meaningful tests that catch real bugs

---

## Testing Pyramid

Our testing strategy follows the testing pyramid:

```
        /\
       /  \      10% - E2E Tests (slow, expensive)
      /____\
     /      \    20% - Integration Tests (moderate speed)
    /________\
   /          \  70% - Unit Tests (fast, cheap)
  /____________\
```

### Test Distribution

| Test Type | Percentage | Count Target | Execution Time |
|-----------|------------|--------------|----------------|
| Unit | 70% | ~1000 tests | <2 minutes |
| Integration | 20% | ~200 tests | <3 minutes |
| E2E/Performance | 10% | ~50 tests | <5 minutes |

---

## Unit Testing

### Framework and Tools

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)
```

### Test Structure

Follow the **Arrange-Act-Assert** pattern:

```go
func TestSearchHandler_Success(t *testing.T) {
    // Arrange - Set up test data and dependencies
    mockRepo := new(MockRepository)
    mockRepo.On("Search", mock.Anything, "query").Return([]Result{
        {ID: "1", Title: "Test"},
    }, nil)
    
    handler := NewSearchHandler(mockRepo)
    req := httptest.NewRequest("GET", "/search?q=query", nil)
    w := httptest.NewRecorder()
    
    // Act - Execute the code under test
    handler.HandleSearch(w, req)
    
    // Assert - Verify expectations
    assert.Equal(t, http.StatusOK, w.Code)
    mockRepo.AssertExpectations(t)
}
```

### Table-Driven Tests

For testing multiple scenarios:

```go
func TestChunking_VariousSizes(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        size     int
        overlap  int
        expected int // expected number of chunks
    }{
        {
            name:     "small document",
            input:    "Hello world",
            size:     100,
            overlap:  20,
            expected: 1,
        },
        {
            name:     "exact fit",
            input:    strings.Repeat("a", 200),
            size:     100,
            overlap:  0,
            expected: 2,
        },
        {
            name:     "with overlap",
            input:    strings.Repeat("a", 200),
            size:     100,
            overlap:  20,
            expected: 3,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            chunker := NewChunker(tt.size, tt.overlap)
            chunks := chunker.Chunk(tt.input)
            assert.Len(t, chunks, tt.expected)
        })
    }
}
```

### Mocking External Dependencies

```go
// Define interface
type EmbeddingService interface {
    Generate(ctx context.Context, text string) ([]float32, error)
}

// Create mock
type MockEmbeddingService struct {
    mock.Mock
}

func (m *MockEmbeddingService) Generate(ctx context.Context, text string) ([]float32, error) {
    args := m.Called(ctx, text)
    return args.Get(0).([]float32), args.Error(1)
}

// Use in tests
func TestIndexer_WithMock(t *testing.T) {
    mockEmbed := new(MockEmbeddingService)
    mockEmbed.On("Generate", mock.Anything, "test").Return(
        []float32{0.1, 0.2, 0.3}, nil,
    )
    
    indexer := NewIndexer(mockEmbed)
    err := indexer.Index(context.Background(), "test")
    
    require.NoError(t, err)
    mockEmbed.AssertExpectations(t)
}
```

### Testing Error Paths

```go
func TestRetrieval_ErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        setupMock     func(*MockVectorDB)
        expectedError string
    }{
        {
            name: "database connection error",
            setupMock: func(m *MockVectorDB) {
                m.On("Search", mock.Anything).Return(
                    nil, errors.New("connection refused"),
                )
            },
            expectedError: "failed to search: connection refused",
        },
        {
            name: "timeout error",
            setupMock: func(m *MockVectorDB) {
                m.On("Search", mock.Anything).Return(
                    nil, context.DeadlineExceeded,
                )
            },
            expectedError: "search timeout",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDB := new(MockVectorDB)
            tt.setupMock(mockDB)
            
            retriever := NewRetriever(mockDB)
            _, err := retriever.Search(context.Background(), "query")
            
            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedError)
        })
    }
}
```

### Running Unit Tests

```bash
# Run all unit tests
go test ./... -short

# Run with coverage
go test ./... -short -cover

# Run with detailed coverage
go test ./... -short -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./internal/retrieval/...

# Run with race detector
go test -race ./...

# Run with verbose output
go test -v ./internal/api/handlers
```

---

## Integration Testing

Integration tests verify that components work together correctly with real dependencies (databases, external APIs, etc.).

### Setup and Teardown

```go
func TestMain(m *testing.M) {
    // Setup
    testDB = setupTestDatabase()
    testQdrant = setupTestQdrant()
    
    // Run tests
    code := m.Run()
    
    // Teardown
    teardownTestDatabase(testDB)
    teardownTestQdrant(testQdrant)
    
    os.Exit(code)
}

func setupTestDatabase() *sql.DB {
    db, err := sql.Open("postgres", getTestDatabaseURL())
    if err != nil {
        log.Fatal(err)
    }
    
    // Run migrations
    runMigrations(db)
    
    return db
}
```

### Database Integration Tests

```go
func TestRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Use test database
    repo := NewRepository(testDB)
    ctx := context.Background()
    
    // Test create
    doc := &Document{
        ID:      "test-1",
        Content: "Integration test document",
    }
    err := repo.Create(ctx, doc)
    require.NoError(t, err)
    
    // Test retrieve
    retrieved, err := repo.Get(ctx, "test-1")
    require.NoError(t, err)
    assert.Equal(t, doc.Content, retrieved.Content)
    
    // Cleanup
    defer repo.Delete(ctx, "test-1")
}
```

### API Integration Tests

```go
func TestSearchAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Start test server
    server := setupTestServer(t)
    defer server.Close()
    
    // Index test document
    indexDoc(t, server.URL, &Document{
        ID:      "test-doc",
        Content: "Go programming language",
    })
    
    // Search for document
    resp := searchRequest(t, server.URL, "golang")
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var results SearchResponse
    json.NewDecoder(resp.Body).Decode(&results)
    assert.Len(t, results.Results, 1)
    assert.Equal(t, "test-doc", results.Results[0].ID)
}
```

### Docker Compose for Integration Tests

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ace_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data  # Use tmpfs for speed

  qdrant-test:
    image: qdrant/qdrant:latest
    ports:
      - "6334:6333"
    tmpfs:
      - /qdrant/storage
```

```bash
# Run integration tests with Docker
docker-compose -f docker-compose.test.yml up -d
go test ./test/integration/... -v
docker-compose -f docker-compose.test.yml down
```

---

## Performance Testing

### Benchmark Tests

```go
func BenchmarkEmbedding_Generate(b *testing.B) {
    service := NewEmbeddingService()
    ctx := context.Background()
    text := "The quick brown fox jumps over the lazy dog"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.Generate(ctx, text)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSearch_Concurrent(b *testing.B) {
    retriever := setupRetriever(b)
    ctx := context.Background()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := retriever.Search(ctx, "test query")
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### Load Testing with k6

```javascript
// k6/load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 users
    { duration: '1m', target: 100 },  // Ramp up to 100 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests < 1s
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  const res = http.get('http://localhost:8080/api/v1/search?q=golang');
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1s': (r) => r.timings.duration < 1000,
  });
  
  sleep(1);
}
```

```bash
# Run load test
k6 run k6/load-test.js

# Run with custom VUs
k6 run --vus 50 --duration 2m k6/load-test.js
```

### Performance Benchmarks

Target metrics for performance tests:

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Simple search | <100ms | P95 latency |
| Complex search (reranking) | <500ms | P95 latency |
| Document indexing | <200ms | P95 latency |
| Embedding generation | <300ms | P95 latency |
| Concurrent searches (100 users) | <1s | P95 latency |

---

## RAG Evaluation Testing

Specialized testing for RAG (Retrieval-Augmented Generation) quality.

### Retrieval Quality Metrics

```go
func TestRetrieval_Quality(t *testing.T) {
    // Load golden dataset
    dataset := loadGoldenDataset("testdata/retrieval-golden.json")
    
    retriever := setupRetriever(t)
    
    var (
        totalRecall    float64
        totalPrecision float64
        totalNDCG      float64
    )
    
    for _, testCase := range dataset {
        results := retriever.Search(context.Background(), testCase.Query)
        
        recall := calculateRecallAtK(results, testCase.RelevantDocs, 10)
        precision := calculatePrecisionAtK(results, testCase.RelevantDocs, 10)
        ndcg := calculateNDCG(results, testCase.RelevantDocs, 10)
        
        totalRecall += recall
        totalPrecision += precision
        totalNDCG += ndcg
    }
    
    avgRecall := totalRecall / float64(len(dataset))
    avgPrecision := totalPrecision / float64(len(dataset))
    avgNDCG := totalNDCG / float64(len(dataset))
    
    // Assert minimum quality thresholds
    assert.GreaterOrEqual(t, avgRecall, 0.8, "Recall@10 should be >= 0.8")
    assert.GreaterOrEqual(t, avgPrecision, 0.7, "Precision@10 should be >= 0.7")
    assert.GreaterOrEqual(t, avgNDCG, 0.75, "NDCG@10 should be >= 0.75")
}

func calculateRecallAtK(results []Result, relevantDocs []string, k int) float64 {
    if len(relevantDocs) == 0 {
        return 0
    }
    
    retrieved := make(map[string]bool)
    for i := 0; i < k && i < len(results); i++ {
        retrieved[results[i].ID] = true
    }
    
    found := 0
    for _, docID := range relevantDocs {
        if retrieved[docID] {
            found++
        }
    }
    
    return float64(found) / float64(len(relevantDocs))
}
```

### Context Relevance Testing

```go
func TestContextRelevance(t *testing.T) {
    tests := []struct {
        query           string
        expectedContext []string // Expected doc IDs in context
        minRelevance    float64  // Minimum relevance score
    }{
        {
            query:           "how to handle errors in go",
            expectedContext: []string{"errors-doc", "best-practices-doc"},
            minRelevance:    0.75,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.query, func(t *testing.T) {
            context := retriever.GetContext(tt.query)
            
            // Check expected docs are present
            for _, expectedID := range tt.expectedContext {
                found := false
                for _, doc := range context.Documents {
                    if doc.ID == expectedID {
                        found = true
                        assert.GreaterOrEqual(t, doc.RelevanceScore, tt.minRelevance)
                    }
                }
                assert.True(t, found, "Expected doc %s not in context", expectedID)
            }
        })
    }
}
```

---

## Test Data Management

### Fixtures

```go
// test/fixtures/documents.go
package fixtures

func NewTestDocument() *Document {
    return &Document{
        ID:       "test-doc-1",
        Title:    "Test Document",
        Content:  "This is test content for unit testing",
        Language: "en",
        Metadata: map[string]string{
            "author": "Test Author",
            "date":   "2025-10-12",
        },
    }
}

func NewTestDocuments(count int) []*Document {
    docs := make([]*Document, count)
    for i := 0; i < count; i++ {
        docs[i] = &Document{
            ID:      fmt.Sprintf("test-doc-%d", i),
            Content: fmt.Sprintf("Test content %d", i),
        }
    }
    return docs
}
```

### Golden Files

```go
func TestParser_GoldenFiles(t *testing.T) {
    files, _ := filepath.Glob("testdata/golden/*.md")
    
    for _, file := range files {
        t.Run(filepath.Base(file), func(t *testing.T) {
            input, _ := os.ReadFile(file)
            expectedOutput, _ := os.ReadFile(strings.Replace(file, ".md", ".json", 1))
            
            parser := NewParser()
            result := parser.Parse(input)
            
            resultJSON, _ := json.MarshalIndent(result, "", "  ")
            
            // Update golden files with -update flag
            if *update {
                os.WriteFile(strings.Replace(file, ".md", ".json", 1), resultJSON, 0644)
            }
            
            assert.JSONEq(t, string(expectedOutput), string(resultJSON))
        })
    }
}
```

### Test Database Seeding

```sql
-- test/fixtures/seed.sql
INSERT INTO documents (id, title, content, created_at) VALUES
('doc-1', 'Go Tutorial', 'Learn Go programming...', NOW()),
('doc-2', 'Python Guide', 'Python best practices...', NOW()),
('doc-3', 'JavaScript Basics', 'JavaScript fundamentals...', NOW());

INSERT INTO chunks (id, document_id, content, embedding) VALUES
('chunk-1', 'doc-1', 'Go is a compiled language', array_fill(0.1, ARRAY[1536])),
('chunk-2', 'doc-1', 'Go has great concurrency', array_fill(0.2, ARRAY[1536]));
```

---

## Coverage Requirements

### Target Coverage Levels

| Component | Unit Coverage | Integration Coverage |
|-----------|---------------|----------------------|
| Core retrieval | 90%+ | 80%+ |
| Indexing pipeline | 85%+ | 70%+ |
| API handlers | 80%+ | 90%+ |
| Storage layer | 85%+ | 85%+ |
| Utilities | 80%+ | N/A |
| **Overall Target** | **85-90%** | **75-80%** |

### Measuring Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage by package
go tool cover -func=coverage.out

# View HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Check coverage threshold
go test ./... -cover | grep -E "coverage: [0-9]+" | \
  awk '{if ($2 < 80) exit 1}'
```

### Coverage Exceptions

Some code is exempt from coverage requirements:

- Generated code (mocks, protobuf)
- Main functions (`cmd/*/main.go`)
- Trivial getters/setters
- Deprecated code marked for removal

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      qdrant:
        image: qdrant/qdrant:latest
        ports:
          - 6333:6333
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run linter
        run: make lint
      
      - name: Run unit tests
        run: go test ./... -short -race -coverprofile=coverage.out
      
      - name: Run integration tests
        run: go test ./test/integration/... -v
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/test?sslmode=disable
          QDRANT_URL: http://localhost:6333
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
      
      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi
```

### Pre-commit Hooks

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Run tests
make test-short || exit 1

# Run linter
make lint || exit 1

# Check formatting
if [ -n "$(gofmt -l .)" ]; then
    echo "Go files are not formatted. Run: make fmt"
    exit 1
fi

echo "All checks passed!"
```

---

## Best Practices

### DO

✅ **Write tests first (TDD)** - Test-driven development catches bugs early  
✅ **Use table-driven tests** - Test multiple scenarios efficiently  
✅ **Test error paths** - Don't just test the happy path  
✅ **Use meaningful test names** - `TestSearchHandler_ReturnsErrorWhenDatabaseDown`  
✅ **Keep tests independent** - Tests should not depend on execution order  
✅ **Use setup/teardown** - Clean up after tests  
✅ **Mock external dependencies** - Don't hit real APIs in tests  
✅ **Test edge cases** - Empty inputs, nil values, boundary conditions  
✅ **Aim for fast tests** - Unit tests should run in milliseconds  
✅ **Use coverage as a guide** - Not a goal, but a signal of gaps

### DON'T

❌ **Don't test implementation details** - Test behavior, not internals  
❌ **Don't write flaky tests** - Tests should be deterministic  
❌ **Don't skip error testing** - Error paths are critical  
❌ **Don't use real databases in unit tests** - Use mocks or in-memory  
❌ **Don't ignore test failures** - Fix or remove broken tests  
❌ **Don't duplicate test logic** - Use helpers and fixtures  
❌ **Don't test external libraries** - Trust third-party code  
❌ **Don't aim for 100% coverage** - Focus on critical paths  
❌ **Don't write tests that test nothing** - Ensure meaningful assertions

### Code Review Checklist

When reviewing tests, check:

- [ ] Tests have clear, descriptive names
- [ ] Both success and error cases are tested
- [ ] Edge cases and boundary conditions are covered
- [ ] Tests are independent and can run in any order
- [ ] Mocks are used appropriately (not over-mocked)
- [ ] Test data is realistic and representative
- [ ] Assertions are meaningful (not just testing type)
- [ ] Test performance is reasonable (<100ms for unit tests)
- [ ] Coverage meets minimum thresholds
- [ ] Tests are maintainable and readable

---

## Quick Reference

### Common Test Commands

```bash
# Run all tests
make test

# Run only unit tests (fast)
make test-unit
go test ./... -short

# Run only integration tests
make test-integration
go test ./test/integration/...

# Run with coverage
make test-coverage

# Run specific test
go test -v -run TestSearchHandler ./internal/api/handlers

# Run benchmarks
go test -bench=. ./internal/retrieval

# Run with race detector
make test-race

# Run with verbose output
go test -v ./...

# Update golden files
go test ./... -update

# Clean test cache
go clean -testcache
```

### Coverage Goals

| Metric | Target | Command |
|--------|--------|---------|
| Overall coverage | 85-90% | `make test-coverage` |
| Unit test coverage | 90%+ | `go test ./... -short -cover` |
| Critical paths | 95%+ | Manual review |

---

**Testing is not about finding bugs; it's about building confidence that our system works as intended.**

For questions or improvements to this strategy, contact the testing team or open an issue.
