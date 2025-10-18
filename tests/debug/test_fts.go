package main

import (
	"database/sql"
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

	// Test FTS search
	query := "Agent"
	rows, err := db.Query(`
		SELECT 
			d.id,
			substr(d.content, 1, 100) as content_preview,
			fts.rank as score
		FROM documents_fts fts
		JOIN documents d ON fts.id = d.id
		WHERE fts.content MATCH ?
		ORDER BY fts.rank ASC
		LIMIT 5
	`, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("FTS search results for '%s':\n", query)
	count := 0
	for rows.Next() {
		var id, contentPreview string
		var score float64
		err = rows.Scan(&id, &contentPreview, &score)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s, Score: %.2f, Content: %s...\n", id, score, contentPreview)
		count++
	}
	fmt.Printf("Total results: %d\n", count)
}
