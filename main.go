package main

import (
	"fmt"
	"nflTerminal/processors"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var json_game_data processors.GameData

type tickMsg time.Time

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

type model struct {
	header       string
	choices      []string         // items on the to-do list
	cursor       int              // which to-do list item our cursor is pointing at
	selected     map[int]struct{} // which to-do items are selected
	footer       string
	tickDuration time.Duration
}

func initialModel() model {
	return model{
		header:       fmt.Sprintf("NFL Games for Week %d, %d\n\n", json_game_data.Week.Number, json_game_data.Season.Year),
		choices:      []string{},
		selected:     make(map[int]struct{}),
		footer:       fmt.Sprintf("\nPress q to quit. Updated: %s\n", time.Now().Format(time.RFC1123)),
		tickDuration: 15 * time.Second,
	}
}

func addChoice(choice string, mod model) model {
	mod.choices = append(mod.choices, choice)
	return mod
}

func updateFooter(mod model) model {
	mod.footer = fmt.Sprintf("\nPress q to quit. Updated: %s\n", time.Now().Format(time.RFC1123))
	return mod
}

func modelTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return modelTickCmd(m.tickDuration)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		json_game_data = processors.GetGameData()
		displayModel := updateModel(m)
		displayModel = updateFooter(displayModel)
		return displayModel, modelTickCmd(m.tickDuration)

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := m.header

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += m.footer

	// Send the UI for rendering
	return s
}

func setModel() model {
	displayModel := initialModel()

	for _, event := range json_game_data.Events {
		for _, competition := range event.Competitions {
			team1 := competition.Competitors[0]
			team2 := competition.Competitors[1]

			displayModel = addChoice(fmt.Sprintf("%3s %s-%s %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation), displayModel)
		}
	}
	return displayModel
}

func updateModel(m model) model {
	m.choices = []string{}

	for _, event := range json_game_data.Events {
		for _, competition := range event.Competitions {
			team1 := competition.Competitors[0]
			team2 := competition.Competitors[1]

			m = addChoice(fmt.Sprintf("%3s %s-%s %s", team1.Team.Abbreviation, team1.Score, team2.Score, team2.Team.Abbreviation), m)
		}
	}
	return m
}
