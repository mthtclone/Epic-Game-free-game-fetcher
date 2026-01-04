package epic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-toast/toast"
)

// Game struct
type Game struct {
	Title       string `json:"title"`
	ProductSlug string `json:"productSlug"`
	Status      string `json:"status"`
	ExpiryDate  string `json:"expiryDate"`
}

// Formatter struct for normalized data
type Formatter struct {
	Title       string
	ProductSlug string
	ExpiryDate  time.Time
	URL         string
}

// Config struct for storing timezone
type Config struct {
	Timezone string `json:"timezone"`
}

var configFile = "config.json"

func SaveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

func LoadConfig() (Config, error) {
	cfg := Config{}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return cfg, nil
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err 
}

// FetchGames fetches the free games from Epic API
func FetchGames() ([]Game, error) {
	url := "https://store-site-backend-static.ak.epicgames.com/freeGamesPromotions"

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "epic-free-game-fetcher/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var epicResp struct {
		Data struct {
			Catalog struct {
				SearchStore struct {
					Elements []Game `json:"elements"`
				} `json:"searchStore"`
			} `json:"Catalog"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&epicResp); err != nil {
		return nil, err
	}

	return epicResp.Data.Catalog.SearchStore.Elements, nil
}

func IsCurrentlyFree(game Game, now time.Time) bool {
	if game.Status != "ACTIVE" || game.ExpiryDate == "" || game.ExpiryDate == "null" {
		return false
	}
	end, err := time.Parse(time.RFC3339, game.ExpiryDate)
	if err != nil {
		fmt.Printf("Warning: could not parse expiryDate for %s: %v\n", game.Title, err)
		return false
	}
	return now.Before(end)
}

// NormalizeData converts raw games to Formatter slice
func NormalizeData(games []Game, now time.Time) []Formatter {
	var result []Formatter
	for _, game := range games {
		if !IsCurrentlyFree(game, now) || game.ProductSlug == "" {
			continue
		}

		end, err := time.Parse(time.RFC3339, game.ExpiryDate)
		if err != nil {
			continue
		}

		result = append(result, Formatter{
			Title:       game.Title,
			ProductSlug: game.ProductSlug,
			ExpiryDate:  end,
			URL:         "https://store.epicgames.com/p/" + game.ProductSlug,
		})
	}
	return result
}

// InferFormat determines format from output file extension or explicit flag
func InferFormat(output, explicitFormat string) string {
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

// WriteFile writes data to a file
func WriteFile(path string, data []byte) {
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Failed to write file %s: %v\n", path, err)
	}
}

// AppendFile appends data to a file
func AppendFile(path string, data []byte) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to write file %s: %v\n", path, err)
		return
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		fmt.Printf("Failed to write file %s: %v\n", path, err)
	}
}

// Notify sends a Windows toast notification
func Notify(title, message string) error {
	notification := toast.Notification{
		AppID:   "Free Epic Game Watcher",
		Title:   title,
		Message: message,
		Icon:    "",
	}
	return notification.Push()
}

// RunAt schedules a function to run daily at a specific time in the set timezone
func RunAt(hour, min int, task func()) {
	go func() {
		for {
			loc := GetTimeZone() 
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))
			task()
		}
	}()
}

// SetTimeZone sets the timezone for scheduled tasks
var location *time.Location = time.UTC
func SetTimeZone(tz string) error {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	location = loc

	cfg := Config{Timezone: tz}
	if err := SaveConfig(cfg); err != nil {
		return err
	}

	return nil
}

// GetTimeZone returns the currently set timezone
func GetTimeZone() *time.Location {
	return location
}