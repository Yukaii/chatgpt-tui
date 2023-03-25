package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"

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

type model struct {
	tabIndex int
	apiKey   string

	// new chatlog should be an array of key/value pairs
	chatLog  []ChatMessage
	textarea textarea.Model

	viewport viewport.Model
	ready    bool
}

// Taken from chatbot-ui
// https://github.com/mckaywrigley/chatbot-ui/blob/main/utils/app/const.ts
const DEFAULT_CHATGPT_PROMPT = "You are ChatGPT, a large language model trained by OpenAI. Follow the user's instructions carefully. Respond using markdown."

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

	chatLog := []ChatMessage{}

	ta := textarea.New()
	ta.Placeholder = "Type your message here..."
	ta.Focus()
	ta.CharLimit = 500
	ta.ShowLineNumbers = false
	ta.Prompt = "┃ "

	return model{
		tabIndex: 0,
		chatLog:  chatLog,
		apiKey:   apiKey,
		textarea: ta,
	}
}

func (m model) Init() bubbletea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var (
		taCmd bubbletea.Cmd
		vpCmd bubbletea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, taCmd = m.textarea.Update(msg)

	var nextCmd bubbletea.Cmd = nil

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		switch msg.Type {

		case bubbletea.KeyCtrlC:
			return m, bubbletea.Quit

		case bubbletea.KeyEnter:
			inputMessage := m.textarea.Value()

			// trim whitespace
			inputMessage = strings.TrimSpace(inputMessage)

			if len(inputMessage) > 0 {
				m.chatLog = append(m.chatLog, ChatMessage{
					role:    "user",
					content: inputMessage,
				})
				m.viewport.SetContent(m.RenderChatLog())
				m.textarea.Reset()

				nextCmd = m.sendMessage(inputMessage)
			}
			break
		}

	case bubbletea.WindowSizeMsg:
		verticalMarginHeight := m.textarea.Height() + 1

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(m.RenderChatLog())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		break

	case ChatMessage:
		m.chatLog = append(m.chatLog, msg)
		m.viewport.SetContent(m.RenderChatLog())
		m.viewport.GotoBottom()
		break
	}

	return m, bubbletea.Batch(taCmd, vpCmd, nextCmd)
}

func (m model) sendMessage(prompt string) bubbletea.Cmd {
	return func() bubbletea.Msg {

		cli, _ := gpt3.NewClient(&gpt3.Options{
			ApiKey:  m.apiKey,
			Timeout: 60 * time.Second,
		})

		messages := []map[string]interface{}{
			{"role": "system", "content": DEFAULT_CHATGPT_PROMPT},
		}

		// append existing messages
		for _, message := range m.chatLog {
			messages = append(messages, map[string]interface{}{
				// set all role to user
				"role":    "user",
				"content": message.content,
			})
		}

		// append new prmopt
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": prompt,
		})

		// Request a chat completion
		uri := "/v1/chat/completions"
		params := map[string]interface{}{
			"model":    "gpt-3.5-turbo",
			"messages": messages,
		}

		res, err := cli.Post(uri, params)
		if err != nil {
			log.Fatalf("request api failed: %v", err)
		}

		message := res.GetString("choices.0.message.content")

		return ChatMessage{
			role:    "ai",
			content: message,
		}
	}
}

var tabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Bottom).Render
var chatUserTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Align(lipgloss.Left).Width(6).Render
var chatAITextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Align(lipgloss.Left).Width(6).Render

func (m model) RenderChatLog() string {
	var maxWidth = m.viewport.Width - 6
	// var chatTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Align(lipgloss.Bottom).MaxWidth(maxWidth).Render
	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(maxWidth),
	)

	chatLogString := ""

	for _, message := range m.chatLog {
		out, _ := r.Render(message.content)

		var who string
		if message.role == "user" {
			who = chatUserTextStyle("You:")
		} else {
			who = chatAITextStyle("AI: ")
		}

		chatLogString += fmt.Sprintf("%s\n%s", who, out)
	}

	return chatLogString
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}
