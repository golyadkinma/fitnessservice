// ai.go — бот-тренер: интеграция внешнего ИИ-сервиса.
//
// ВНИМАНИЕ: в константе aiAPIKey оставлено МЕСТО ДЛЯ КЛЮЧА API.
// Вставьте действующий ключ сервиса (например, OpenAI-совместимого) в aiAPIKey,
// при необходимости поправьте aiEndpoint и модель в aiModel.
// Пока ключ пуст, бот отвечает в демо-режиме на локальных правилах.
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// === ДЫРКА ДЛЯ КЛЮЧА API ==============================================
// Вставьте ключ сюда:
const aiAPIKey = "" // например: "sk-..."

// Параметры внешнего ИИ-сервиса (OpenAI-совместимый протокол).
const (
	aiEndpoint = "https://api.openai.com/v1/chat/completions"
	aiModel    = "gpt-4o-mini"
	aiTimeout  = 30 * time.Second
)
// ======================================================================

// AIRequest — запрос пользователя боту-тренеру.
type AIRequest struct {
	Message string `json:"message"` // вопрос пользователя
	Weight  float64 `json:"weight"` // масса пользователя, кг
	Height  float64 `json:"height"` // рост пользователя, м
	Steps   int    `json:"steps"`   // шаги за сегодня
}

// AIResponse — ответ бота-тренера.
type AIResponse struct {
	Reply  string `json:"reply"`
	Source string `json:"source"` // "ai" — внешний сервис, "demo" — локальные правила
}

// AskAI задаёт вопрос боту-тренеру.
// При наличии ключа API запрос уходит во внешний сервис,
// иначе используется демо-режим на локальных правилах.
func AskAI(req AIRequest) (AIResponse, error) {
	question := strings.TrimSpace(req.Message)
	if question == "" {
		return AIResponse{}, errors.New("вопрос не может быть пустым")
	}
	if len(question) > 1000 {
		return AIResponse{}, errors.New("вопрос слишком длинный, до 1000 символов")
	}
	if aiAPIKey == "" {
		return AIResponse{Reply: demoReply(question, req), Source: "demo"}, nil
	}
	reply, err := remoteAI(question, req)
	if err != nil {
		// Внешний сервис недоступен — не прерываем работу приложения.
		return AIResponse{Reply: demoReply(question, req) + " (внешний сервис недоступен)", Source: "demo"}, nil
	}
	return AIResponse{Reply: reply, Source: "ai"}, nil
}

// remoteAI выполняет запрос к внешнему ИИ-сервису.
func remoteAI(question string, req AIRequest) (string, error) {
	payload := map[string]any{
		"model": aiModel,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": fmt.Sprintf(
					"Ты бот-тренер фитнес-приложения. Отвечай кратко и по делу, на русском языке. "+
						"Данные пользователя: масса %.1f кг, рост %.2f м, шагов сегодня %d.",
					req.Weight, req.Height, req.Steps),
			},
			{"role": "user", "content": question},
		},
		"max_tokens": 500,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequest(http.MethodPost, aiEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+aiAPIKey)
	client := &http.Client{Timeout: aiTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("сервис вернул статус %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("пустой ответ сервиса")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// demoReply — демо-режим: локальные правила на основе данных пользователя.
func demoReply(question string, req AIRequest) string {
	q := strings.ToLower(question)
	bmiRes, bmiErr := CalcBMI(req.Weight, req.Height)
	switch {
	case strings.Contains(q, "имт") || strings.Contains(q, "масс"):
		if bmiErr != nil {
			return "Заполните профиль (масса и рост), и я рассчитаю индекс массы тела."
		}
		return fmt.Sprintf("Ваш ИМТ %.1f — %s. %s.", bmiRes.Value, bmiRes.Category, bmiRes.Advice)
	case strings.Contains(q, "шаг"):
		if req.Steps >= 8000 {
			return fmt.Sprintf("Отлично, сегодня уже %d шагов — дневная норма выполнена! Поддерживайте темп.", req.Steps)
		}
		return fmt.Sprintf("Сегодня %d шагов. До нормы 8000 осталось %d — короткая прогулка закроет план.", req.Steps, 8000-req.Steps)
	case strings.Contains(q, "питан") || strings.Contains(q, "рацион") || strings.Contains(q, "протеин"):
		return "Основной приём белка распределяйте на 3–4 порции по 25–30 г. Каталог спортивного питания — во вкладке «Каталог»."
	case strings.Contains(q, "тренировк") || strings.Contains(q, "упражн"):
		return "Начните с разминки 10 минут, основная часть 30–40 минут, заминка 5 минут. Силовые — 2–3 раза в неделю."
	case strings.Contains(q, "вода") || strings.Contains(q, "пить"):
		return "Норма воды — примерно 30 мл на килограмм массы в день, на тренировке добавьте 500 мл."
	default:
		return "Я демо-бот-тренер. Спросите про ИМТ, шаги, питание, тренировки или питьевой режим. " +
			"Для полноценных ответов добавьте ключ API в server/ai.go."
	}
}
