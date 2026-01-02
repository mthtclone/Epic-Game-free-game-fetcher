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

	f := epic.InferFormat(*output, *format)

	var data []byte

	switch f {
	case "text":
		fmt.Println("Formatting in Text")
		data = []byte(FormatText(freeGames))
	case "json":
		fmt.Println("Formatting in JSON")
		var err error
		data, err = FormatJSON(freeGames)
		if err != nil {
			fmt.Printf("Failed to format in JSON: %v\n", err)
			data = nil
		}
	case "html":
		fmt.Println("Formatting in HTML")
		htmlOut, err := FormatHTML(freeGames)
		if err != nil {
			fmt.Printf("Failed to format in HTML: %v\n", err)
			data = nil
		}
		data = []byte(htmlOut)
	default:
		fmt.Println("Unknown format, defaulting to text")
		data = []byte(FormatText(freeGames))
	}

	if *output != "" {
		if *appendMode {
			epic.AppendFile(*output, data)
		} else {
			epic.WriteFile(*output, data)
		}
	} else {
		fmt.Print(string(data))
	}
}
