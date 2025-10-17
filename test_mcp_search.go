package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
)

func main() {
	// Open the same database
	store, err := sqlite.NewStore("data/conexus.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Create mock embedder
	embedder := embedding.NewMock(768)

	// Test the search
	ctx := context.Background()
	query := "package"
	
	// Generate embedding
	emb, err := embedder.Embed(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	// Call SearchHybrid
	opts := vectorstore.SearchOptions{
		Limit: 5,
	}
	results, err := store.SearchHybrid(ctx, query, emb.Vector, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Search results for '%s': %d results\n", query, len(results))
	for i, result := range results {
		fmt.Printf("%d. ID: %s, Score: %.3f, Content preview: %.50s...\n", 
			i+1, result.Document.ID, result.Score, result.Document.Content)
	}
}
