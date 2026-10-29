// steps.go — шагомер: журнал шагов по дням, статистика и серия дней с выполненной нормой.
// Шагомер в мини-приложении работает как заглушка: данные вносятся вручную
// или подтягиваются демо-значениями, интеграция с системными сенсорами планируется отдельно.
package server

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// WeeklyStats — сводка активности за неделю.
type WeeklyStats struct {
	Total      int     `json:"total"`      // шагов за неделю
	Average    int     `json:"average"`    // среднее шагов в день
	DistanceKm float64 `json:"distanceKm"` // дистанция, км
	Kcal       int     `json:"kcal"`       // расход, ккал
	Streak     int     `json:"streak"`     // серия дней с выполненной нормой
	Goal       int     `json:"goal"`       // дневная норма шагов
	Days       []StepDay `json:"days"`     // записи по дням
}

const (
	stepsPerKm     = 1350 // среднее число шагов в километре
	kcalPer1000    = 40   // средний расход ккал на 1000 шагов
	defaultGoal    = 8000
	minStepsPerDay = 0
	maxStepsPerDay = 200000
)

// AddSteps сохраняет количество шагов за указанный день.
// Повторная запись за тот же день заменяет предыдущую.
func AddSteps(entries []StepDay, date string, steps int) ([]StepDay, error) {
	if _, err := time.Parse(dateLayout, date); err != nil {
		return nil, fmt.Errorf("дата должна быть в формате ГГГГ-ММ-ДД: %w", err)
	}
	if steps < minStepsPerDay || steps > maxStepsPerDay {
		return nil, fmt.Errorf("количество шагов должно быть от %d до %d", minStepsPerDay, maxStepsPerDay)
	}
	for i, e := range entries {
		if e.Date == date {
			entries[i].Steps = steps
			return entries, nil
		}
	}
	return append(entries, StepDay{Date: date, Steps: steps}), nil
}

// SortSteps упорядочивает записи по возрастанию даты.
func SortSteps(entries []StepDay) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Date < entries[j].Date })
}

// StepsOn возвращает количество шагов за конкретный день (0, если записи нет).
func StepsOn(entries []StepDay, date string) int {
	for _, e := range entries {
		if e.Date == date {
			return e.Steps
		}
	}
	return 0
}

// CurrentStreak возвращает длину серии дней подряд, в которые выполнена норма.
// Отсчёт ведётся от указанной даты назад. Текущий день ещё не завершён,
// поэтому невыполненная норма сегодня не обрывает серию.
func CurrentStreak(entries []StepDay, today string, goal int) int {
	cursor, err := time.Parse(dateLayout, today)
	if err != nil {
		return 0
	}
	streak := 0
	first := true
	for {
		key := cursor.Format(dateLayout)
		steps := StepsOn(entries, key)
		if steps < goal {
			if first && steps < goal {
				// Сегодняшний день не завершён — не обрывает серию.
				first = false
				cursor = cursor.AddDate(0, 0, -1)
				continue
			}
			break
		}
		first = false
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

// WeeklyStats возвращает сводку активности за последние 7 дней от указанной даты.
func Weekly(entries []StepDay, today string, goal int) (WeeklyStats, error) {
	end, err := time.Parse(dateLayout, today)
	if err != nil {
		return WeeklyStats{}, errors.New("некорректная дата")
	}
	SortSteps(entries)
	days := make([]StepDay, 0, 7)
	total := 0
	for i := 6; i >= 0; i-- {
		key := end.AddDate(0, 0, -i).Format(dateLayout)
		steps := StepsOn(entries, key)
		days = append(days, StepDay{Date: key, Steps: steps})
		total += steps
	}
	streak := CurrentStreak(entries, today, goal)
	return WeeklyStats{
		Total:      total,
		Average:    total / 7,
		DistanceKm: round2(float64(total) / stepsPerKm),
		Kcal:       total / 1000 * kcalPer1000,
		Streak:     streak,
		Goal:       goal,
		Days:       days,
	}, nil
}
