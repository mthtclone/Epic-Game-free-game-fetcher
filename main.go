package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"flag"
	"os"
	"strings"
	"path/filepath"
	"github.com/go-toast/toast"
)

type EpicResponse struct {
	Data struct {
		Catalog struct {
			SearchStore struct {
				Elements []Game `json:"elements"`
			} `json:"searchStore"`
		} `json:"Catalog"`
	} `json:"data"`
}

type Game struct {
	Title       string     `json:"title"`
	ProductSlug string     `json:"productSlug"`
	Status      string     `json:"status"`
	ExpiryDate  string     `json:"expiryDate"`
}

func main() {

	format := flag.String("format", "text", "Output format: text, json, html")
	output := flag.String("output", "", "Output file path")
	appendMode := flag.Bool("append", false, "Append to existing file")
	// stateFile := flag.String("state", "state.json", "Path to state JSON file")

	flag.Parse()

	var games []Game
	var err error

	for i := 0; i < 3; i ++ {
		games, err = fetchGames()
		if err == nil {
			break
		}

		fmt.Printf("Fetch failed, retrying in 5s... (%v)\n", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		fmt.Println("Failed to fetch games after 3 attempts:", err)
		return
	}

	freeGames := normalizeData(games, time.Now().UTC())
	f := inferFormat(*output, *format)

	if len(freeGames) == 0 {
		notify("No Free Games to claim today!", "")
	} else {
		for _, game := range freeGames {
			err := notify("Free Game Available!", game.Title + " is free until "+ game.ExpiryDate.Format("Mon Jan 2 15:04:05 2006 MST"))
			if err != nil {
				fmt.Printf("Failed to send notification for %s: %v\n", game.Title, err)
			}
		}
	}
	
	var data []byte

	switch f {
	case "text":
		fmt.Print("Formatting in Text\n")
		data = []byte(FormatText(freeGames))

	case "json":
		fmt.Print("Formatting in JSON\n")
		var err error
		data, err = FormatJSON(freeGames)
		if err != nil {
			fmt.Printf("Failed to format in JSON: %v\n", err)
        	data = nil
		}

	case "html":
		fmt.Print("Formatting in HTML\n")
		htmlOut, err := FormatHTML(freeGames)
		if err != nil {
			fmt.Printf("Failed to format in HTML: %v\n", err)
        	data = nil
		}
		data = []byte(htmlOut)

	default:
		panic("Unknown format, please specify format: text, json, or html")
		data = nil
	}

	if *output != "" {
		if *appendMode {
			appendFile(*output, data)
		} else {
			writeFile(*output, data)
		} 
		if err != nil {
			fmt.Printf("Failed to write to file %s: %v\n", *output, err)
		}
	} else {
		fmt.Print(string(data))
	}
 }

func fetchGames() ([]Game, error) {
	url := "https://store-site-backend-static.ak.epicgames.com/freeGamesPromotions"

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	req.Header.Set("User-Agent", "epic-free-game-fetcher/1.0")

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var epic EpicResponse
	if err := json.NewDecoder(resp.Body).Decode(&epic); err != nil {
		return nil, err
	}

	return epic.Data.Catalog.SearchStore.Elements, nil
}

func isCurrentlyFree(game Game, now time.Time) bool {
	if game.Status != "ACTIVE" {
		return false
	}

	if game.ExpiryDate == "" || game.ExpiryDate == "null" {
		return false
	}

	end, err := time.Parse(time.RFC3339, game.ExpiryDate) 
	if err != nil {
		fmt.Printf("Warning: could not parse expiryDate for %s: %v\n", game.Title, err)
		return false
	}

	return now.Before(end)
}

func normalizeData(games []Game, now time.Time) []Formatter {
	var result []Formatter
	for _, game := range games {
		if !isCurrentlyFree(game, now) || game.ProductSlug == "" {
			continue
		}

		end, err := time.Parse(time.RFC3339, game.ExpiryDate)
		if err != nil {
			continue
		}

		result = append(result, Formatter{
			Title:			game.Title,
			ProductSlug:	game.ProductSlug,
			ExpiryDate:		end,
			URL:         	"https://store.epicgames.com/p/" + game.ProductSlug,			
		})
	}

	return result
}

func inferFormat(output, explicitFormat string) string {
	if explicitFormat != "" {
		return strings.ToLower(explicitFormat)
	}

	if output != "" {
		ext := strings.ToLower(filepath.Ext(output))
		switch ext {
		case ".json":
			return "json"
		case ".html", ".htm":
			return "html"
		case ".txt":
			return "text"
		}
	}

	return "text"
}

func writeFile(path string, data []byte) {
	if err := os.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
}

func appendFile(path string, data []byte) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		panic(err)
	}
}

func notify(title, message string) error {
	notification := toast.Notification{
		AppID:		"Free Epic Game Watcher",
		Title:		title,
		Message: 	message,
		Icon:		"",
	}

	return notification.Push()
}