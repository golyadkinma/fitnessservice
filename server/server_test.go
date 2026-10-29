// server_test.go — комплект модульных тестов серверной части FitTrack.
// Тесты не используют внешних библиотек: только стандартный пакет testing.
package server

import (
	"strings"
	"testing"
	"time"
)

// --- Калькулятор ИМТ -------------------------------------------------------

func TestBMINormal(t *testing.T) {
	res, err := CalcBMI(70, 1.75)
	if err != nil {
		t.Fatalf("не ожидалась ошибка: %v", err)
	}
	if res.Value != 22.86 {
		t.Fatalf("ИМТ: ожидалось 22.86, получено %.2f", res.Value)
	}
	if res.Category != "Нормальная масса тела" {
		t.Fatalf("категория: ожидалась норма, получено %q", res.Category)
	}
}

func TestBMIBoundaries(t *testing.T) {
	cases := []struct {
		weight, height float64
		want           string
	}{
		{55, 1.8, "Недостаточная масса тела"},
		{45, 1.7, "Выраженный дефицит массы тела"},
		{76, 1.74, "Избыточная масса тела"},
		{95, 1.75, "Ожирение I степени"},
		{115, 1.7, "Ожирение II степени"},
		{130, 1.7, "Ожирение III степени"},
	}
	for _, c := range cases {
		res, err := CalcBMI(c.weight, c.height)
		if err != nil {
			t.Fatalf("%f/%f: ошибка: %v", c.weight, c.height, err)
		}
		if res.Category != c.want {
			t.Fatalf("%f/%f: ожидалась категория %q, получено %q", c.weight, c.height, c.want, res.Category)
		}
	}
}

func TestBMIInvalidInput(t *testing.T) {
	if _, err := CalcBMI(10, 1.75); err == nil {
		t.Fatal("ожидалась ошибка для массы ниже допустимой")
	}
	if _, err := CalcBMI(70, 5.0); err == nil {
		t.Fatal("ожидалась ошибка для роста выше допустимого")
	}
}

// --- Шагомер ---------------------------------------------------------------

func TestAddStepsReplacesSameDay(t *testing.T) {
	entries, err := AddSteps(nil, "2026-10-28", 5000)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	entries, err = AddSteps(entries, "2026-10-28", 7000)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if len(entries) != 1 || entries[0].Steps != 7000 {
		t.Fatalf("ожидалась одна запись с 7000 шагов, получено %+v", entries)
	}
}

func TestAddStepsValidation(t *testing.T) {
	if _, err := AddSteps(nil, "28.10.2026", 100); err == nil {
		t.Fatal("ожидалась ошибка для неверного формата даты")
	}
	if _, err := AddSteps(nil, "2026-10-28", -5); err == nil {
		t.Fatal("ожидалась ошибка для отрицательных шагов")
	}
}

func TestStreakCountsFromToday(t *testing.T) {
	// Данные внесены только до вчерашнего дня — серия не должна «зависать».
	entries := []StepDay{
		{Date: "2026-10-26", Steps: 9000},
		{Date: "2026-10-27", Steps: 9000},
	}
	if got := CurrentStreak(entries, "2026-10-28", 8000); got != 2 {
		t.Fatalf("серия: ожидалось 2 (сегодня не завершён), получено %d", got)
	}
	// Сегодняшний день выполнен — серия растёт.
	entries, err := AddSteps(entries, "2026-10-28", 10000)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if got := CurrentStreak(entries, "2026-10-28", 8000); got != 3 {
		t.Fatalf("серия: ожидалось 3, получено %d", got)
	}
	// Вчерашний день провален — серия обрывается на сегодня.
	entries, err = AddSteps(entries, "2026-10-27", 100)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if got := CurrentStreak(entries, "2026-10-28", 8000); got != 1 {
		t.Fatalf("серия: ожидался 1, получено %d", got)
	}
}

func TestWeeklyStats(t *testing.T) {
	entries := []StepDay{
		{Date: "2026-10-22", Steps: 9000},
		{Date: "2026-10-25", Steps: 8000},
		{Date: "2026-10-26", Steps: 8000},
		{Date: "2026-10-27", Steps: 8000},
	}
	stats, err := Weekly(entries, "2026-10-28", 8000)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if stats.Total != 33000 {
		t.Fatalf("итого: ожидалось 33000, получено %d", stats.Total)
	}
	if stats.Streak != 3 {
		t.Fatalf("серия: ожидалось 3, получено %d", stats.Streak)
	}
	if len(stats.Days) != 7 {
		t.Fatalf("дней в сводке: ожидалось 7, получено %d", len(stats.Days))
	}
}

// --- Геймификация ----------------------------------------------------------

func TestLevels(t *testing.T) {
	if LevelFor(0) != 1 || LevelFor(499) != 1 || LevelFor(500) != 2 || LevelFor(1200) != 3 {
		t.Fatal("некорректный расчёт уровней")
	}
	if LevelProgress(550) != 10 {
		t.Fatalf("прогресс: ожидалось 10, получено %d", LevelProgress(550))
	}
}

func TestUpdateByStepsGrantsAchievements(t *testing.T) {
	p := UpdateBySteps(UserProgress{}, map[string]int{"2026-10-28": 21000}, 3)
	if !has(p.Achievements, "first_steps") || !has(p.Achievements, "walker") ||
		!has(p.Achievements, "goalkeeper") || !has(p.Achievements, "marathon") {
		t.Fatalf("ожидались достижения за шаги, получено %v", p.Achievements)
	}
	if !has(p.Achievements, "streak_3") {
		t.Fatalf("ожидалось достижение streak_3, получено %v", p.Achievements)
	}
	if !has(p.Achievements, "streak_7") {
		t.Logf("достижение streak_7 корректно не выдано при серии 3")
	} else {
		t.Fatal("достижение streak_7 не должно выдаваться при серии 3")
	}
}

// --- Каталог ---------------------------------------------------------------

func TestCatalogFilter(t *testing.T) {
	products := defaultProducts()
	got := FilterProducts(products, "протеин", "")
	if len(got) != 3 { // протеин Whey, протеин Casein и протеиновый батончик
		t.Fatalf("поиск «протеин»: ожидалось 3 товара, получено %d", len(got))
	}
	got = FilterProducts(products, "", "изотоник")
	if len(got) != 2 {
		t.Fatalf("категория «изотоник»: ожидалось 2 товара, получено %d", len(got))
	}
	got = FilterProducts(products, "", "")
	if len(got) != len(products) {
		t.Fatalf("без условий: ожидалось %d товаров, получено %d", len(products), len(got))
	}
	if got[0].ID != 1 {
		t.Fatal("в наличии должны идти раньше отсутствующих")
	}
}

func TestCatalogCategories(t *testing.T) {
	cats := Categories(defaultProducts())
	if len(cats) == 0 || !strings.Contains(strings.Join(cats, ","), "протеин") {
		t.Fatalf("категории определены некорректно: %v", cats)
	}
}

// --- Челленджи -------------------------------------------------------------

func TestChallengePace(t *testing.T) {
	c := Challenge{Code: "test", Target: 56000, StartDate: "2026-10-26", EndDate: "2026-11-01"}
	pace := PaceFor(c, "u1", map[string]int{"2026-10-27": 8000}, "2026-10-28")
	if pace.Remaining != 48000 {
		t.Fatalf("остаток: ожидалось 48000, получено %d", pace.Remaining)
	}
	if pace.PerDay != 9600 { // 48000 / 5 дней, округление вверх
		t.Fatalf("темп: ожидалось 9600, получено %d", pace.PerDay)
	}
	if pace.DonePct != 14.29 {
		t.Fatalf("процент: ожидалось 14.29, получено %.2f", pace.DonePct)
	}
}

func TestChallengeValidation(t *testing.T) {
	valid := Challenge{Code: "x", Target: 100, StartDate: "2026-11-01", EndDate: "2026-11-07"}
	if err := ValidateChallenge(valid); err != nil {
		t.Fatalf("валидный челлендж отклонён: %v", err)
	}
	bad := valid
	bad.Target = 0
	if err := ValidateChallenge(bad); err == nil {
		t.Fatal("ожидалась ошибка для нулевой цели")
	}
	bad2 := valid
	bad2.EndDate = "2026-10-01"
	if err := ValidateChallenge(bad2); err == nil {
		t.Fatal("ожидалась ошибка при дате окончания раньше начала")
	}
}

// --- Напоминания -----------------------------------------------------------

func TestReminderValidation(t *testing.T) {
	valid := Reminder{Title: "Тренировка", Time: "19:30", Days: []int{1, 3, 5}}
	if err := ValidateReminder(valid); err != nil {
		t.Fatalf("валидное напоминание отклонено: %v", err)
	}
	bad := valid
	bad.Time = "25:00"
	if err := ValidateReminder(bad); err == nil {
		t.Fatal("ожидалась ошибка для времени 25:00")
	}
	bad2 := valid
	bad2.Days = nil
	if err := ValidateReminder(bad2); err == nil {
		t.Fatal("ожидалась ошибка без дней недели")
	}
}

func TestSortDays(t *testing.T) {
	got := SortDays([]int{5, 1, 3, 1})
	if len(got) != 3 || got[0] != 1 || got[1] != 3 || got[2] != 5 {
		t.Fatalf("ожидался [1 3 5], получено %v", got)
	}
}

// --- Чат -------------------------------------------------------------------

func TestChatHistoryLimit(t *testing.T) {
	var messages []ChatMessage
	for i := 1; i <= MaxChatHistory+10; i++ {
		messages = AppendChat(messages, i, "Аня", "сообщение")
	}
	if len(messages) != MaxChatHistory {
		t.Fatalf("ожидалось %d сообщений, получено %d", MaxChatHistory, len(messages))
	}
	if messages[len(messages)-1].ID != MaxChatHistory+10 {
		t.Fatal("последнее сообщение должно быть самым свежим")
	}
}

func TestChatEmptyMessageIgnored(t *testing.T) {
	before := time.Now()
	messages := AppendChat(nil, 1, "Аня", "   ")
	if len(messages) != 0 {
		t.Fatal("пустое сообщение не должно добавляться")
	}
	_ = before
}

// --- Хранилище -------------------------------------------------------------

func TestStoreRoundTrip(t *testing.T) {
	path := t.TempDir() + "/state.json"
	s1, err := NewStore(path)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if err := s1.Seed(); err != nil {
		t.Fatalf("ошибка наполнения: %v", err)
	}
	err = s1.Update(func(a *AppState) error {
		a.Steps["u1"] = []StepDay{{Date: "2026-10-28", Steps: 12345}}
		return nil
	})
	if err != nil {
		t.Fatalf("ошибка сохранения: %v", err)
	}
	s2, err := NewStore(path)
	if err != nil {
		t.Fatalf("ошибка повторной загрузки: %v", err)
	}
	s2.View(func(a *AppState) {
		if len(a.Steps["u1"]) != 1 || a.Steps["u1"][0].Steps != 12345 {
			t.Fatalf("состояние не совпало после перезагрузки: %+v", a.Steps["u1"])
		}
		if len(a.Products) != len(defaultProducts()) {
			t.Fatal("каталог не сохранился")
		}
	})
}

func TestStoreRollbackOnError(t *testing.T) {
	path := t.TempDir() + "/state.json"
	s, _ := NewStore(path)
	_ = s.Seed()
	err := s.Update(func(a *AppState) error {
		a.Products = nil
		return errNotFound
	})
	if err == nil {
		t.Fatal("ожидалась ошибка")
	}
	s.View(func(a *AppState) {
		if len(a.Products) == 0 {
			t.Fatal("изменение должно быть откачено при ошибке")
		}
	})
}
