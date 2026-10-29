// challenges.go — челленджи: групповые цели с расчётом дневного темпа.
package server

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ValidateChallenge проверяет корректность параметров челленджа.
func ValidateChallenge(c Challenge) error {
	if strings.TrimSpace(c.Code) == "" {
		return errors.New("код челленджа не может быть пустым")
	}
	if c.Target <= 0 {
		return fmt.Errorf("цель должна быть положительной, получено %d", c.Target)
	}
	start, err := time.Parse(dateLayout, c.StartDate)
	if err != nil {
		return errors.New("некорректная дата начала")
	}
	end, err := time.Parse(dateLayout, c.EndDate)
	if err != nil {
		return errors.New("некорректная дата окончания")
	}
	if end.Before(start) {
		return errors.New("дата окончания раньше даты начала")
	}
	return nil
}

// PaceFor рассчитывает дневной темп достижения цели участником.
// stepsByDay — шаги участника за дни периода челленджа.
func PaceFor(c Challenge, userID string, stepsByDay map[string]int, today string) ChallengePace {
	start, _ := time.Parse(dateLayout, c.StartDate)
	end, _ := time.Parse(dateLayout, c.EndDate)
	now, err := time.Parse(dateLayout, today)
	if err != nil || now.Before(start) {
		now = start
	}

	done := 0
	for key, v := range stepsByDay {
		day, err := time.Parse(dateLayout, key)
		if err != nil {
			continue
		}
		if !day.Before(start) && !day.After(end) {
			done += v
		}
	}

	daysLeft := int(end.Sub(now).Hours()/24) + 1
	if daysLeft < 0 {
		daysLeft = 0
	}
	remaining := c.Target - done
	if remaining < 0 {
		remaining = 0
	}
	perDay := 0
	if daysLeft > 0 {
		perDay = (remaining + daysLeft - 1) / daysLeft // округление вверх
	}
	donePct := 0.0
	if c.Target > 0 {
		donePct = round2(float64(done) / float64(c.Target) * 100)
	}
	return ChallengePace{
		DaysLeft:  daysLeft,
		Remaining: remaining,
		PerDay:    perDay,
		DonePct:   donePct,
	}
}
