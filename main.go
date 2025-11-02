package main

import (
	"fmt"
	"nflTerminal/processors"
	"os"
	"strings"
	"time"

	models "nflTerminal/models"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
)

type model models.Model
type tickMsg time.Time

var json_game_data processors.GameData
var lastPlay string

var styleHeader = gloss.NewStyle().Bold(true).Underline(true).Foreground(gloss.Color("#00719b"))
var styleInprogress = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ffffff"))
var styleScheduled = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#808080"))
var styleFinal = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#bd0000"))
var styleAdditionalInfo = gloss.NewStyle().Italic(true).Foreground(gloss.Color("#ffffff"))
var styleFooter = gloss.NewStyle().Italic(true).Foreground(gloss.Color("#808080"))

func main() {
	// TODO: GET nfl games for the week
	json_game_data = processors.GetGameData()

	if len(json_game_data.Events) == 0 {
		fmt.Println("No game data available. Exiting.")
		return
	}

	// TODO: Update model with each game

	displayModel := setModel()

	p := tea.NewProgram(displayModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

func initialModel() model {
	return model{
		Header:         styleHeader.Render(fmt.Sprintf("NFL Games for Week %d, %d", json_game_data.Week.Number, json_game_data.Season.Year)),
		Choices:        []string{},
		Selected:       int(-1),
		AdditionalInfo: styleAdditionalInfo.Render("\n\nMake a selection for additional info.\n"),
		Footer:         styleFooter.Render(fmt.Sprintf("\nPress q to quit. Updated: %s\n", time.Now().Format("15:04:05"))),
		TickDuration:   15 * time.Second,
	}
}

func updateFooter(mod model) model {
	mod.Footer = styleFooter.Render(fmt.Sprintf("\nPress q to quit. Updated: %s\n", time.Now().Format("15:04:05")))
	return mod
}

func modelTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return modelTickCmd(m.TickDuration)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		json_game_data = processors.GetGameData()
		displayModel := updateModel(m)
		displayModel = updateFooter(displayModel)

		if m.Selected != -1 {
			displayModel.AdditionalInfo = styleAdditionalInfo.Render(fmt.Sprintf("\n\n%s\n", updateLastPlay(m.Selected)))
		}

		return displayModel, modelTickCmd(m.TickDuration)

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the Cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// The "down" and "j" keys move the Cursor down
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the Selected state for the item that the Cursor is pointing at.
		case "enter", " ":
			m.Selected = m.Cursor

			m.AdditionalInfo = styleAdditionalInfo.Render(fmt.Sprintf("\n\n%s\n", updateLastPlay(m.Selected)))
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The Header
	s := m.Header + "\n"

	// Iterate over our Choices
	for i, choice := range m.Choices {

		// Is the Cursor pointing at this choice?
		Cursor := " " // no Cursor
		if m.Cursor == i {
			Cursor = ">" // Cursor!
		}

		// Is this choice Selected?
		checked := " " // not Selected
		if m.Selected == i {
			checked = "x" // Selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", Cursor, checked, choice)
	}

	// The Additional Information
	s += m.AdditionalInfo

	// The Footer
	s += m.Footer

	// Send the UI for rendering
	return s
}

func setModel() model {
	displayModel := initialModel()

	displayModel = generateChoices(displayModel)

	return displayModel
}

func updateModel(m model) model {
	m.Choices = []string{}

	m = generateChoices(m)

	return m
}

func generateChoices(m model) model {
	scheduled := []string{}
	inprogress := []string{}
	final := []string{}

	for _, event := range json_game_data.Events {
		competition := event.Competitions[0]
		team1 := competition.Competitors[0]
		team2 := competition.Competitors[1]

		switch event.Status.Type.Name {
		case "STATUS_FINAL":
			final = append(final, styleFinal.Render(fmt.Sprintf("%3s %-2s-%2s %3s: %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation, event.Status.Type.ShortDetail)))
		case "STATUS_IN_PROGRESS":
			inprogress = append(inprogress, styleInprogress.Render(fmt.Sprintf("%3s %-2s-%2s %3s: %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation, event.Status.Type.ShortDetail)))
		case "STATUS_SCHEDULED":
			scheduled = append(scheduled, styleScheduled.Render(fmt.Sprintf("%3s %-2s-%2s %3s: %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation, event.Status.Type.ShortDetail)))
		default:
			inprogress = append(inprogress, styleInprogress.Render(fmt.Sprintf("%3s %-2s-%2s %3s: %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation, event.Status.Type.ShortDetail)))
		}

	}

	m.Choices = append(m.Choices, inprogress...)
	m.Choices = append(m.Choices, scheduled...)
	m.Choices = append(m.Choices, final...)

	return m
}

func updateLastPlay(selection int) string {
	var lastPlayTeamAbbre string

	competition := json_game_data.Events[selection].Competitions[0]

	lastPlay = competition.Situation.LastPlay.Text
	lastPlayTeamID := competition.Situation.LastPlay.Team.ID

	if lastPlayTeamID == competition.Competitors[0].ID {
		lastPlayTeamAbbre = competition.Competitors[0].Team.Abbreviation
	} else if lastPlayTeamID == competition.Competitors[1].ID {
		lastPlayTeamAbbre = competition.Competitors[1].Team.Abbreviation
	}

	if lastPlay != strings.TrimSpace("") {
		lastPlay = fmt.Sprintf("(%s): %s", lastPlayTeamAbbre, lastPlay)
	} else {
		lastPlay = "No recent play information available."
	}

	return competition.Situation.LastPlay.Text
}
