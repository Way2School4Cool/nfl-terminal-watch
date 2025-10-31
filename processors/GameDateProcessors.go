package processors

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type GameData struct {
	// Define fields based on the expected JSON structure
	Season struct {
		Year int `json:"year"`
	} `json:"season"`
	Week struct {
		Number int `json:"number"`
	} `json:"week"`
	Events []struct {
		Name         string `json:"name"`
		Competitions []struct {
			Competitors []struct {
				Team struct {
					Abbreviation string `json:"abbreviation"`
					DisplayName  string `json:"displayName"`
				} `json:"team"`
				HomeAway string `json:"homeAway"`
				Score    string `json:"score"`
			} `json:"competitors"`
		} `json:"competitions"`
		Status struct {
			Type struct {
				Detail string `json:"detail"`
			} `json:"type"`
		} `json:"status"`
	} `json:"events"`
}

func GetGameData() GameData {
	// Try the current week's scoreboard
	resp, err := http.Get("https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard")
	if err != nil {
		log.Fatalln("Could not get active NFL data:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Could not read NFL data response:", err)
	}

	// Check for error response
	var errorResp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Code == 404 {
		log.Println("Error: API endpoint not found. The NFL season might be in offseason or the endpoint has changed.")
		return GameData{}
	}

	jsonData := GameData{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Fatalln("Could not unmarshal NFL data response:", err)
	}

	// Verify the structure matches
	if len(jsonData.Events) == 0 {
		log.Println("Warning: No events found in the response")
	}

	/*
		log.Printf("Successfully retrieved NFL data for season %d week %d", jsonData.Season.Year, jsonData.Week.Number)
		for _, event := range jsonData.Events {
			log.Printf("Event: %s", event.Name)
			for _, competition := range event.Competitions {
				for _, competitor := range competition.Competitors {
					log.Printf("%s (%s): %s", competitor.Team.DisplayName, competitor.Team.Abbreviation, competitor.Score)
				}
			}
			log.Printf("Status: %s", event.Status.Type.Detail)
		}
	*/

	return jsonData

}
