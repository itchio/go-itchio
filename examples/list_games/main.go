package main

import (
	"context"
	"fmt"
	"os"

	itchio "github.com/itchio/go-itchio"
)

func main() {
	apiKey := os.Getenv("ITCHIO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ITCHIO_API_KEY environment variable is required")
		os.Exit(1)
	}

	client := itchio.ClientWithKey(apiKey)
	ctx := context.Background()

	resp, err := client.ListProfileGames(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d games:\n\n", len(resp.Games))
	for _, game := range resp.Games {
		fmt.Printf("  [%d] %s\n", game.ID, game.Title)
		if game.URL != "" {
			fmt.Printf("       %s\n", game.URL)
		}
		fmt.Println()
	}
}
