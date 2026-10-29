// reminders.go — напоминания по дням недели.
package server

import (
	"errors"
	"fmt"
	"strings"
)

// weekdayNames — человекочитаемые названия дней недели.
var weekdayNames = []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"}

// ValidateReminder проверяет корректность напоминания.
func ValidateReminder(r Reminder) error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("название напоминания не может быть пустым")
	}
	if len(r.Time) != 5 || r.Time[2] != ':' {
		return errors.New("время должно быть в формате ЧЧ:ММ")
	}
	hh, mm := 0, 0
	if _, err := fmt.Sscanf(r.Time, "%2d:%2d", &hh, &mm); err != nil || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return errors.New("время должно быть в формате ЧЧ:ММ")
	}
	if len(r.Days) == 0 {
		return errors.New("укажите хотя бы один день недели")
	}
	for _, d := range r.Days {
		if d < 0 || d > 6 {
			return errors.New("день недели должен быть числом от 0 до 6")
		}
	}
	return nil
}

// DaysString возвращает дни недели в виде строки «пн, ср, пт».
func DaysString(days []int) string {
	parts := make([]string, 0, len(days))
	for _, d := range days {
		if d >= 0 && d <= 6 {
			parts = append(parts, weekdayNames[d])
		}
	}
	return strings.Join(parts, ", ")
}

// SortDays возвращает отсортированный список дней недели без повторов.
func SortDays(days []int) []int {
	seen := map[int]bool{}
	result := []int{}
	for _, d := range days {
		if !seen[d] {
			seen[d] = true
			result = append(result, d)
		}
	}
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j] < result[j-1]; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}
	return result
}
