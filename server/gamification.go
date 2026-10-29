// gamification.go — подсистема геймификации: очки, уровни, достижения.
package server

import "fmt"

// Thresholds уровня: очки от начала уровня до следующего.
const pointsPerLevel = 500

// achievements — каталог достижений приложения.
var achievements = []Achievement{
	{"first_steps", "Первые шаги", 50, 1000},
	{"walker", "Ходок", 100, 5000},
	{"goalkeeper", "Норма дня", 150, 8000},
	{"marathon", "Марафонец", 300, 20000},
	{"streak_3", "Три дня подряд", 200, 8000},
	{"streak_7", "Неделя без пропусков", 500, 8000},
}

// LevelFor вычисляет уровень по общему числу очков (уровень = очки/500 + 1).
func LevelFor(points int) int {
	return points/pointsPerLevel + 1
}

// LevelProgress возвращает процент прогресса внутри текущего уровня.
func LevelProgress(points int) int {
	return points % pointsPerLevel * 100 / pointsPerLevel
}

// AddPoints начисляет очки и пересчитывает уровень.
func AddPoints(p UserProgress, delta int) UserProgress {
	p.Points += delta
	p.Level = LevelFor(p.Points)
	p.LevelProgress = LevelProgress(p.Points)
	return p
}

// UpdateBySteps пересчитывает очки и достижения по журналам шагов.
// stepsByDay — суммарные шаги по дням, streak — текущая серия.
func UpdateBySteps(p UserProgress, stepsByDay map[string]int, streak int) UserProgress {
	var maxDay int
	for _, v := range stepsByDay {
		if v > maxDay {
			maxDay = v
		}
	}
	for _, a := range achievements {
		if has(p.Achievements, a.Code) {
			continue
		}
		granted := false
		switch a.Code {
		case "first_steps", "walker", "goalkeeper", "marathon":
			granted = maxDay >= a.MinSteps
		case "streak_3":
			granted = streak >= 3
		case "streak_7":
			granted = streak >= 7
		}
		if granted {
			p.Achievements = append(p.Achievements, a.Code)
			p = AddPoints(p, a.Points)
		}
	}
	return p
}

// AchievementName возвращает название достижения по коду.
func AchievementName(code string) string {
	for _, a := range achievements {
		if a.Code == code {
			return a.Name
		}
	}
	return code
}

// Achievements возвращает каталог достижений.
func Achievements() []Achievement { return achievements }

func has(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// ValidateProgress проверяет корректность состояния прогресса.
func ValidateProgress(p UserProgress) error {
	if p.Points < 0 {
		return fmt.Errorf("очки не могут быть отрицательными")
	}
	return nil
}
