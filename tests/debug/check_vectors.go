package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/conexus.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check vector data
	rows, err := db.Query("SELECT id, substr(vector, 1, 50) as vector_preview FROM documents LIMIT 3")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Vector data samples:")
	for rows.Next() {
		var id, vectorPreview string
		err = rows.Scan(&id, &vectorPreview)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s, Vector preview: %s...\n", id, vectorPreview)
	}

	// Try to parse a vector
	var vectorJSON string
	err = db.QueryRow("SELECT vector FROM documents LIMIT 1").Scan(&vectorJSON)
	if err != nil {
		log.Fatal(err)
	}

	var vector []float32
	err = json.Unmarshal([]byte(vectorJSON), &vector)
	if err != nil {
		log.Printf("Failed to parse vector JSON: %v", err)
	} else {
		fmt.Printf("Parsed vector length: %d, first 5 values: %v\n", len(vector), vector[:min(5, len(vector))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
