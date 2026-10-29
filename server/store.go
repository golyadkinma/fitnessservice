// Package server — серверная часть фитнес-приложения FitTrack.
// store.go — единое файловое хранилище состояния приложения в формате JSON.
package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// AppState — полное состояние приложения, сохраняемое на диск.
type AppState struct {
	Products    []Product            `json:"products"`
	Steps       map[string][]StepDay `json:"steps"`       // ключ — идентификатор пользователя
	Streak      map[string]int       `json:"streak"`      // серия дней по пользователям
	Reminders   map[string][]Reminder `json:"reminders"`
	Challenges  []Challenge          `json:"challenges"`
	ChallengeJoins map[string][]string `json:"challengeJoins"` // пользователь -> коды челленджей
	Progress    map[string]UserProgress `json:"progress"`    // пользователь -> очки и уровень
	Chat        []ChatMessage        `json:"chat"`
	Profiles    map[string]Profile   `json:"profiles"`
}

// Store — потокобезопасное хранилище с сохранением в JSON-файл.
type Store struct {
	mu       sync.Mutex
	path     string
	state    AppState
}

// NewStore создаёт хранилище и загружает состояние из файла, если он существует.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	s.state = defaultState()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil // первый запуск — используем состояние по умолчанию
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &s.state); err != nil {
		return nil, err
	}
	return s, nil
}

// defaultState возвращает начальное состояние с каталогом продукции по умолчанию.
func defaultState() AppState {
	return AppState{
		Products: defaultProducts(),
		Steps:    map[string][]StepDay{},
		Streak:   map[string]int{},
		Reminders: map[string][]Reminder{},
		ChallengeJoins: map[string][]string{},
		Progress: map[string]UserProgress{},
		Profiles: map[string]Profile{},
	}
}

// Save записывает состояние на диск атомарно: сначала во временный файл.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Seed наполняет состояние стартовыми данными при первом запуске.
func (s *Store) Seed() error {
	return s.Update(func(a *AppState) error {
		if len(a.Challenges) == 0 {
			a.Challenges = defaultChallenges()
		}
		if len(a.Products) == 0 {
			a.Products = defaultProducts()
		}
		return nil
	})
}

// View выполняет функцию чтения под блокировкой.
func (s *Store) View(fn func(*AppState)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.state)
}

// Update выполняет функцию изменения под блокировкой и сохраняет состояние.
// Если функция возвращает ошибку, изменение откатывается и состояние не сохраняется.
func (s *Store) Update(fn func(*AppState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	backup := s.state
	if err := fn(&s.state); err != nil {
		s.state = backup
		return err
	}
	return s.saveLocked()
}
