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

	// Check for Go files
	rows, err := db.Query("SELECT id, substr(content, 1, 200) as content_preview, metadata FROM documents WHERE metadata LIKE '%go%' LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Go files in database:")
	count := 0
	for rows.Next() {
		count++
		var id, contentPreview, metadata string
		err = rows.Scan(&id, &contentPreview, &metadata)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("Content preview: %s...\n", contentPreview)
		fmt.Printf("Metadata: %s\n", metadata)
		fmt.Println("---")
	}
	
	if count == 0 {
		fmt.Println("No Go files found in database!")
		
		// Check total count
		var total int
		err = db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&total)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Total documents: %d\n", total)
	}
}
