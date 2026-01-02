package main

import (
	"fmt"
	"time"
	"flag"
	"free-game-fetcher-bot/pkg/epic" 
)

func main() {

	format := flag.String("format", "text", "Output format: text, json, html")
	output := flag.String("output", "", "Output file path")
	appendMode := flag.Bool("append", false, "Append to existing file")
	flag.Parse()

	games, err := epic.FetchGames()
	if err != nil {
		fmt.Println("Failed to fetch games:", err)
		return
	}

	freeGames := epic.NormalizeData(games, time.Now().UTC())
	fmt.Println(freeGames)
}
