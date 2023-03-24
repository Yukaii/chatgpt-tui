package main

import (
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	bubbletea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/chatgp/gpt3"
	"github.com/joho/godotenv"
	"github.com/muesli/termenv"
	"os"
)

type ChatMessage struct {
	role    string
	content string
}

type (
	errMsg error
)

type model struct {
	tabIndex int
	apiKey   string

	// new chatlog should be an array of key/value pairs
	chatLog   []ChatMessage
	textarea textarea.Model
	err       errMsg
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	termenv.HideCursor()
	defer termenv.ShowCursor()

	p := bubbletea.NewProgram(initialModel(), bubbletea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialModel() model {
	apiKey := os.Getenv("OPENAI_API_KEY")

	var message = "Welcome to the chat!"

	chatLog := []ChatMessage{
		{
			role:    "AI",
			content: message,
		},
	}

	ta := textarea.New()
	ta.Placeholder = "Type your message here..."
	ta.Focus()
	ta.CharLimit = 500
	ta.ShowLineNumbers = false

	return model{
		tabIndex:  0,
		chatLog:   chatLog,
		apiKey:    apiKey,
		textarea:  ta,
	}
}

func (m model) Init() bubbletea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var taCmd bubbletea.Cmd

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		switch msg.Type {

		case bubbletea.KeyCtrlC, bubbletea.KeyEsc:
			return m, bubbletea.Quit

		case bubbletea.KeyEnter:
			var inputMessage = m.textarea.Value()
			if len(inputMessage) > 0 {
				m.chatLog = append(m.chatLog, ChatMessage{
					role:    "user",
					content: inputMessage,
				})
				m.textarea.SetValue("")
				return m, m.sendMessage()
			}
		}
	case errMsg:
		m.err = msg
		return m, nil

	case ChatMessage:
		m.chatLog = append(m.chatLog, msg)
		return m, nil
	}

	m.textarea, taCmd = m.textarea.Update(msg)

	return m, taCmd
}

func (m model) sendMessage() bubbletea.Cmd {
	return func() bubbletea.Msg {
		// Set up the GPT-3 client
		cli, _ := gpt3.NewClient(&gpt3.Options{
			ApiKey:  m.apiKey,
			Timeout: 30 * time.Second,
		})

		// Request a chat completion
		uri := "/v1/chat/completions"
		params := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{"role": "user", "content": m.textarea.Value()},
			},
		}

		res, err := cli.Post(uri, params)
		if err != nil {
			log.Fatalf("request api failed: %v", err)
		}

		message := res.GetString("choices.0.message.content")

		return ChatMessage{
			role:    "AI",
			content: message,
		}
	}
}

var tabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Bottom).Render
var chatTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Align(lipgloss.Bottom).Render

func (m model) View() string {
	view := strings.Builder{}

	var chatLog = ""
	for _, message := range m.chatLog {
		chatLog += message.role + ": " + message.content + "\n"
	}

	view.WriteString(chatLog)

	view.WriteString("\n")
	view.WriteString(m.textarea.View())

	// Render help text
	helpText := "Press TAB to switch between input and buttons. Press ENTER to send a message. Press ESC or CTRL+C to exit."
	view.WriteString("\n\n" + tabStyle(helpText))

	return view.String()
}
