package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	itchio "github.com/itchio/go-itchio"
)

func main() {
	apiKey := os.Getenv("ITCHIO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ITCHIO_API_KEY environment variable is required")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: search_games <query>")
		fmt.Fprintln(os.Stderr, "Example: search_games \"platformer\"")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")

	client := itchio.ClientWithKey(apiKey)
	ctx := context.Background()

	resp, err := client.SearchGames(ctx, itchio.SearchGamesParams{
		Query: query,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Search results for %q (page %d, %d per page):\n\n", query, resp.Page, resp.PerPage)

	if len(resp.Games) == 0 {
		fmt.Println("No games found.")
		return
	}

	for _, game := range resp.Games {
		fmt.Printf("  [%d] %s", game.ID, game.Title)
		if game.Classification != "" {
			fmt.Printf(" (%s)", game.Classification)
		}
		fmt.Println()

		if game.ShortText != "" {
			fmt.Printf("       %s\n", game.ShortText)
		}
		if game.URL != "" {
			fmt.Printf("       %s\n", game.URL)
		}
		fmt.Println()
	}
}
