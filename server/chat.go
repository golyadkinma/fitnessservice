// chat.go — внутренний чат команды.
package server

import (
	"strings"
	"time"
)

// MaxChatHistory — сколько последних сообщений возвращает сервер.
const MaxChatHistory = 100

// MaxChatTextLen — максимальная длина сообщения.
const MaxChatTextLen = 500

// AppendChat добавляет сообщение в чат, обрезая историю до MaxChatHistory.
func AppendChat(messages []ChatMessage, id int, user, text string) []ChatMessage {
	text = strings.TrimSpace(text)
	if text == "" || len(text) > MaxChatTextLen {
		return messages
	}
	if user == "" {
		user = "Гость"
	}
	messages = append(messages, ChatMessage{
		ID:   id,
		User: user,
		Text: text,
		At:   time.Now().Format(time.RFC3339),
	})
	if len(messages) > MaxChatHistory {
		messages = messages[len(messages)-MaxChatHistory:]
	}
	return messages
}
