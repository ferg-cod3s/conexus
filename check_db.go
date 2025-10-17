package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/conexus.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check documents table
	rows, err := db.Query("SELECT id, substr(content, 1, 200) as content_preview, metadata FROM documents LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("First 5 documents:")
	for rows.Next() {
		var id, contentPreview, metadata string
		err = rows.Scan(&id, &contentPreview, &metadata)
		if err != nil {
			log.Fatal(err)
		}
		
		// Check if content looks like actual code
		isCode := strings.Contains(contentPreview, "func") || 
		          strings.Contains(contentPreview, "package") ||
		          strings.Contains(contentPreview, "import")
		
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("Content preview: %s...\n", contentPreview)
		fmt.Printf("Looks like code: %v\n", isCode)
		fmt.Printf("Metadata: %s\n", metadata)
		fmt.Println("---")
	}
}
