package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	Promotions  Promotions `json:"promotions"`
	ExpiryDate  string     `json:"expiryDate"`
}

func main() {

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
	
	now := time.Now().UTC()
	fmt.Println("Currently Free Epic Games:\n")

	for _, game := range games {
		if isCurrentlyFree(game, now) && game.ProductSlug != "" {
			fmt.Println(game.Title)
			fmt.Println("https://store.epicgames.com/p/" + game.ProductSlug)
			fmt.Println()
		}
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