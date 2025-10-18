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

	// Check all files
	rows, err := db.Query("SELECT metadata FROM documents ")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Files in database:")
	for rows.Next() {
		var metadataStr string
		err = rows.Scan(&metadataStr)
		if err != nil {
			log.Fatal(err)
		}
		
		var metadata map[string]interface{}
		err = json.Unmarshal([]byte(metadataStr), &metadata)
		if err != nil {
			fmt.Printf("Error parsing metadata: %v\n", err)
			continue
		}
		
		filePath, ok := metadata["file_path"].(string)
		if !ok {
			fmt.Printf("No file_path in metadata: %s\n", metadataStr)
			continue
		}
		
		fmt.Printf("File: %s\n", filePath)
	}
}
