package main

import (
	"fmt"
	"log"
	"time"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/muesli/reflow/wordwrap"

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

	chatLog := []ChatMessage{
		{
			role:    "AI",
			content: "Welcome to the chat!",
		},
	}

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

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		switch msg.Type {

		case bubbletea.KeyCtrlC, bubbletea.KeyEsc:
			return m, bubbletea.Quit

		case bubbletea.KeyEnter:
			inputMessage := m.textarea.Value()

			// trim whitespace
			inputMessage = strings.TrimSpace(inputMessage)

			if len(inputMessage) > 0 {
				const aiText = "Hello world"

				m.chatLog = append(m.chatLog, ChatMessage{
					role:    "user",
					content: inputMessage,
				}, ChatMessage{
					role:    "AI",
					content: aiText,
				})

				m.textarea.Reset()

				m.viewport.SetContent(m.RenderChatLog())
				m.viewport.GotoBottom()
			}
		}

	case bubbletea.WindowSizeMsg:
		verticalMarginHeight := m.textarea.Height() + 2

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(m.RenderChatLog())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	return m, bubbletea.Batch(taCmd, vpCmd)
}

func (m model) CallChatGPTApi() string {
	cli, _ := gpt3.NewClient(&gpt3.Options{
		ApiKey:  m.apiKey,
		Timeout: 30 * time.Second,
	})

	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello world"},
	}

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

	return message
}

// func (m model) sendMessage() bubbletea.Cmd {
// 	return func() bubbletea.Msg {
// 		// Set up the GPT-3 client
//
// 		return ChatMessage{
// 			role:    "AI",
// 			content: message,
// 		}
// 	}
// }

var tabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Bottom).Render
var chatTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Align(lipgloss.Bottom).Render

func (m model) RenderChatLog() string {
	chatLogString := ""

	for _, message := range m.chatLog {
		chatLogString += message.role + ": " + message.content + "\n"
	}

	return chatLogString
}

func (m model) View() string {
	return fmt.Sprintf(
		"%d\n%s\n\n%s\n%s",
		len(m.chatLog),
		m.viewport.View(),
		m.textarea.View(),
		wordwrap.String(tabStyle("Press TAB to switch between input and buttons. Press ENTER to send a message. Press ESC or CTRL+C to exit."), m.viewport.Width),
	)

	// view := strings.Builder{}
	//
	// var chatLog = ""
	// for _, message := range m.chatLog {
	// 	chatLog += message.role + ": " + message.content + "\n"
	// }
	//
	// view.WriteString(chatLog)
	//
	// view.WriteString("\n")
	// view.WriteString(m.textarea.View())
	//
	// // Render help text
	// helpText := "Press TAB to switch between input and buttons. Press ENTER to send a message. Press ESC or CTRL+C to exit."
	// view.WriteString("\n\n" + tabStyle(helpText))
	//
	// return view.String()
}
