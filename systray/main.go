package main 

import (
	"fmt"
	"time"
	_ "time/tzdata"
	_ "embed"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"github.com/mthtclone/Epic-Game-free-game-fetcher/pkg/epic"
)

//go:embed icon.ico
var iconData []byte

func main () {
	
	cfg, err := epic.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
	} else if cfg.Timezone != "" {
		_ = epic.SetTimeZone(cfg.Timezone)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Free Epic Game Fetcher")
	systray.SetTooltip("Claim free games from Epic Games store on time.")
	systray.SetIcon(iconData)

	mRefresh := systray.AddMenuItem("Refresh Now", "Check for free games now.")
	mSetTime := systray.AddMenuItem("Adjust timezone", "Adjust time according to your time-zone.")
	mQuit := systray.AddMenuItem("Quit", "Click to disable service.")

	epic.RunAt(22, 30, func() {
		games, err := epic.FetchGames()
		if err != nil {
			fmt.Println("Failed to fetch games: %v\n", err)
			return
		}
		freeGames := epic.NormalizeData(games, time.Now().UTC())
		if len(freeGames) == 0 {
			epic.Notify("No Free Games", "")
		} else {
			for _, g := range freeGames {
				epic.Notify("Free Game Available!", g.Title+" until "+g.ExpiryDate.Format("Mon Jan 2 15:04 MST"))
			}
		}
	})

	go func() {
		for {
			select {
			case <-mRefresh.ClickedCh:
				fmt.Println("Refreshing games...")
				games, err := epic.FetchGames()
				if err != nil {
					epic.Notify("Fail to fetch games.", "")
					continue
				}

				freeGames := epic.NormalizeData(games, time.Now().UTC())
				if len(freeGames) == 0 {
					epic.Notify("No Free Games", "")
				} else {
					for _, g := range freeGames {
						epic.Notify("Free Game Available!", g.Title+" until "+g.ExpiryDate.Format("Mon Jan 2 15:04 MST"))
					}
				}
			
			case <-mSetTime.ClickedCh:
				fmt.Println("Adjusting Time")
				currentTZ := epic.GetTimeZone().String()

				zenity.Info(
					"Currently set timezone:\n\n"+currentTZ,
					zenity.Title("Current Timezone"),
				)

				input, err := zenity.Entry(
					"Enter your timezone (-> Asia/Tokyo):", 
					zenity.Title("Set Timezone"),
				)
				
				if err != nil {
					fmt.Printf("Timezone input canceled or error: %v\n", err)
					epic.Notify("Input Canceled", input)
					break
				}

				if err := epic.SetTimeZone(input); err != nil {
					fmt.Printf("Failed to set timezone: %v\n", err)
					epic.Notify("Fail to update Timezone", input)
				} else {
					updatedTZ := epic.GetTimeZone().String()

					zenity.Info(
						"Updated Timezone:\n\n"+updatedTZ,
						zenity.Title("Updated Timezone"),
					)

					fmt.Printf("Timezone set to %s\n", input)
					epic.Notify("Timezone Updated", "Timezone set to "+input)
				}

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	fmt.Println("Exiting...")
}

// build cmd
// cd systray
// set G0111MODULE=on
// go build -ldflags "-H windowsgui" -o freeEpicWatcher.exe main.go