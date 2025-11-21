package main

import (
	"fmt"
	"strings"
)

func parseRepository(repo string) (owner, name string) {
	parts := strings.Split(repo, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", repo
}

func main() {
	owner, repo := parseRepository("ferg-cod3s/conexus")
	fmt.Printf("Owner: %s, Repo: %s\n", owner, repo)
}
