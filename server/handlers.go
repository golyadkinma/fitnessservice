// handlers.go — REST API приложения: маршруты, разбор запросов, идентификация пользователя.
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RegisterRoutes регистрирует все маршруты API и раздаёт клиентскую часть.
func RegisterRoutes(mux *http.ServeMux, store *Store, webDir string) {
	// Раздача клиентской части мини-приложения.
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/catalog", getCatalog(store))
	mux.HandleFunc("POST /api/bmi", postBMI)
	mux.HandleFunc("GET /api/steps", getSteps(store))
	mux.HandleFunc("POST /api/steps", postSteps(store))
	mux.HandleFunc("GET /api/reminders", getReminders(store))
	mux.HandleFunc("POST /api/reminders", postReminders(store))
	mux.HandleFunc("DELETE /api/reminders/{id}", deleteReminder(store))
	mux.HandleFunc("GET /api/challenges", getChallenges(store))
	mux.HandleFunc("POST /api/challenges/{code}/join", joinChallenge(store))
	mux.HandleFunc("GET /api/chat", getChat(store))
	mux.HandleFunc("POST /api/chat", postChat(store))
	mux.HandleFunc("POST /api/ai/chat", postAI(store))
	mux.HandleFunc("GET /api/profile", getProfile(store))
	mux.HandleFunc("POST /api/profile", postProfile(store))
}

// userID определяет пользователя запроса по initData Telegram.
func userID(r *http.Request) string {
	user, err := ValidateInitData(r.Header.Get("X-Telegram-Init-Data"))
	if err != nil || user.ID == 0 {
		return "guest"
	}
	return "tg-" + strconv.FormatInt(user.ID, 10)
}

// userName возвращает отображаемое имя пользователя.
func userName(r *http.Request) string {
	user, err := ValidateInitData(r.Header.Get("X-Telegram-Init-Data"))
	if err != nil || user.FirstName == "" {
		return ""
	}
	return user.FirstName
}

// writeJSON отправляет ответ в формате JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// readJSON разбирает тело запроса в формате JSON.
func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON: " + err.Error()})
		return false
	}
	return true
}

// fail отправляет сообщение об ошибке.
func fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// health — проверка живости сервера.
func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
}

// getCatalog возвращает каталог с поиском и фильтром по категории.
func getCatalog(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var products []Product
		store.View(func(s *AppState) { products = s.Products })
		writeJSON(w, http.StatusOK, map[string]any{
			"products":   FilterProducts(products, r.URL.Query().Get("q"), r.URL.Query().Get("category")),
			"categories": Categories(products),
		})
	}
}

// postBMI рассчитывает индекс массы тела.
func postBMI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Weight float64 `json:"weight"`
		Height float64 `json:"height"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	res, err := CalcBMI(req.Weight, req.Height)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// getSteps возвращает недельную статистику шагов.
func getSteps(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userID(r)
		goal := defaultGoal
		if g := r.URL.Query().Get("goal"); g != "" {
			if v, err := strconv.Atoi(g); err == nil && v >= 1000 && v <= 50000 {
				goal = v
			}
		}
		var entries []StepDay
		store.View(func(s *AppState) {
			entries = append([]StepDay(nil), s.Steps[uid]...)
		})
		stats, err := Weekly(entries, time.Now().Format(dateLayout), goal)
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, stats)
	}
}

// postSteps сохраняет количество шагов за день.
func postSteps(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StepDay
		if !readJSON(w, r, &req) {
			return
		}
		uid := userID(r)
		err := store.Update(func(s *AppState) error {
			updated, err := AddSteps(s.Steps[uid], req.Date, req.Steps)
			if err != nil {
				return err
			}
			SortSteps(updated)
			s.Steps[uid] = updated
			return nil
		})
		if err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// getReminders возвращает напоминания пользователя.
func getReminders(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userID(r)
		var list []Reminder
		store.View(func(s *AppState) { list = s.Reminders[uid] })
		if list == nil {
			list = []Reminder{}
		}
		writeJSON(w, http.StatusOK, list)
	}
}

// postReminders создаёт напоминание.
func postReminders(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Reminder
		if !readJSON(w, r, &req) {
			return
		}
		req.Days = SortDays(req.Days)
		req.Active = true
		uid := userID(r)
		var id int
		err := store.Update(func(s *AppState) error {
			if err := ValidateReminder(req); err != nil {
				return err
			}
			id = nextID(s.Reminders[uid], func(x Reminder) int { return x.ID })
			req.ID = id
			s.Reminders[uid] = append(s.Reminders[uid], req)
			return nil
		})
		if err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, req)
	}
}

// deleteReminder удаляет напоминание по идентификатору.
func deleteReminder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fail(w, http.StatusBadRequest, "некорректный идентификатор")
			return
		}
		uid := userID(r)
		err = store.Update(func(s *AppState) error {
			list := s.Reminders[uid]
			for i, item := range list {
				if item.ID == id {
					s.Reminders[uid] = append(list[:i], list[i+1:]...)
					return nil
				}
			}
			return errNotFound
		})
		if err != nil {
			fail(w, http.StatusNotFound, "напоминание не найдено")
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

var errNotFound = errors.New("не найдено")

// getChallenges возвращает челленджи с прогрессом пользователя.
func getChallenges(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userID(r)
		type joined struct {
			Challenge
			Joined  bool          `json:"joined"`
			Pace    ChallengePace `json:"pace"`
			Members int           `json:"members"`
		}
		result := []joined{}
		store.View(func(s *AppState) {
			steps := map[string]int{}
			for _, d := range s.Steps[uid] {
				steps[d.Date] = d.Steps
			}
			for _, c := range s.Challenges {
				result = append(result, joined{
					Challenge: c,
					Joined:  contains(s.ChallengeJoins[uid], c.Code),
					Pace:    PaceFor(c, uid, steps, time.Now().Format(dateLayout)),
					Members: countMembers(s.ChallengeJoins, c.Code),
				})
			}
		})
		writeJSON(w, http.StatusOK, result)
	}
}

// joinChallenge регистрирует пользователя в челлендже.
func joinChallenge(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		uid := userID(r)
		err := store.Update(func(s *AppState) error {
			found := false
			for _, c := range s.Challenges {
				if strings.EqualFold(c.Code, code) {
					found = true
					break
				}
			}
			if !found {
				return errNotFound
			}
			if contains(s.ChallengeJoins[uid], code) {
				return nil
			}
			s.ChallengeJoins[uid] = append(s.ChallengeJoins[uid], code)
			return nil
		})
		if err != nil {
			fail(w, http.StatusNotFound, "челлендж не найден")
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// getChat возвращает историю чата.
func getChat(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var messages []ChatMessage
		store.View(func(s *AppState) {
			start := 0
			if len(s.Chat) > MaxChatHistory {
				start = len(s.Chat) - MaxChatHistory
			}
			messages = append([]ChatMessage(nil), s.Chat[start:]...)
		})
		if messages == nil {
			messages = []ChatMessage{}
		}
		writeJSON(w, http.StatusOK, messages)
	}
}

// postChat добавляет сообщение в чат.
func postChat(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Text string `json:"text"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Text) == "" {
			fail(w, http.StatusBadRequest, "сообщение не может быть пустым")
			return
		}
		name := userName(r)
		if name == "" {
			name = "Гость"
		}
		err := store.Update(func(s *AppState) error {
			id := nextID(s.Chat, func(m ChatMessage) int { return m.ID })
			s.Chat = AppendChat(s.Chat, id, name, req.Text)
			return nil
		})
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// postAI передаёт вопрос боту-тренеру.
func postAI(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AIRequest
		if !readJSON(w, r, &req) {
			return
		}
		uid := userID(r)
		store.View(func(s *AppState) {
			if p, ok := s.Profiles[uid]; ok && req.Weight == 0 && req.Height == 0 {
				req.Weight = p.Weight
				req.Height = p.Height
			}
		})
		today := time.Now().Format(dateLayout)
		store.View(func(s *AppState) {
			for _, d := range s.Steps[uid] {
				if d.Date == today {
					req.Steps = d.Steps
				}
			}
		})
		res, err := AskAI(req)
		if err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

// getProfile возвращает профиль с прогрессом геймификации.
func getProfile(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := userID(r)
		var profile Profile
		var progress UserProgress
		store.View(func(s *AppState) {
			profile = s.Profiles[uid]
			progress = s.Progress[uid]
		})
		names := map[string]string{}
		for _, a := range Achievements() {
			names[a.Code] = a.Name
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"profile":      profile,
			"progress":     progress,
			"achievements": names,
			"catalog":      Achievements(),
		})
	}
}

// postProfile сохраняет анкету пользователя и пересчитывает прогресс геймификации.
func postProfile(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Profile
		if !readJSON(w, r, &req) {
			return
		}
		if req.Goal < 1000 || req.Goal > 50000 {
			req.Goal = defaultGoal
		}
		if req.Weight != 0 && req.Height != 0 {
			if _, err := CalcBMI(req.Weight, req.Height); err != nil {
				fail(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		uid := userID(r)
		if p := userName(r); p != "" && req.Name == "" {
			req.Name = p
		}
		err := store.Update(func(s *AppState) error {
			s.Profiles[uid] = req
			// Пересчёт очков и достижений по журналу шагов.
			stepsByDay := map[string]int{}
			for _, d := range s.Steps[uid] {
				stepsByDay[d.Date] = d.Steps
			}
			today := time.Now().Format(dateLayout)
			s.Streak[uid] = CurrentStreak(s.Steps[uid], today, req.Goal)
			s.Progress[uid] = UpdateBySteps(s.Progress[uid], stepsByDay, s.Streak[uid])
			return nil
		})
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// nextID возвращает следующий свободный идентификатор списка.
func nextID[T any](list []T, key func(T) int) int {
	max := 0
	for _, item := range list {
		if key(item) > max {
			max = key(item)
		}
	}
	return max + 1
}

// contains проверяет наличие строки в списке.
func contains(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}

// countMembers считает участников челленджа.
func countMembers(joins map[string][]string, code string) int {
	n := 0
	for _, codes := range joins {
		if contains(codes, code) {
			n++
		}
	}
	return n
}
