// telegram.go — интеграция с Telegram: проверка подписи initData мини-приложения
// и заглушка бота для напоминаний.
package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
)

// === МЕСТО ДЛЯ ТОКЕНА БОТА ============================================
// Вставьте токен бота (от @BotFather) в botToken при развёртывании напоминаний в Telegram.
const botToken = ""
// ======================================================================

// TelegramUser — пользователь, переданный Telegram в initData мини-приложения.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// ValidateInitData проверяет подпись initData от Telegram WebApp.
// Возвращает идентификатор пользователя. Пока токен бота не задан,
// проверка пропускается и возвращается гостевой идентификатор.
func ValidateInitData(initData string) (TelegramUser, error) {
	if botToken == "" || initData == "" {
		// Демо-режим вне Telegram: работает браузерная версия.
		return TelegramUser{ID: 0, FirstName: "Гость"}, nil
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, errors.New("некорректный initData")
	}
	var user TelegramUser
	if raw := values.Get("user"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &user); err != nil {
			return TelegramUser{}, errors.New("некорректные данные пользователя")
		}
	}
	hash := values.Get("hash")
	values.Del("hash")
	// Сортировка полей и сбор строки проверки.
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(values.Get(k))
	}
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sb.String()))
	if hex.EncodeToString(mac.Sum(nil)) != hash {
		return TelegramUser{}, errors.New("подпись initData не совпала")
	}
	return user, nil
}

// hmacSHA256 вычисляет HMAC-SHA256 с ключом key над сообщением message.
func hmacSHA256(key, message []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}
