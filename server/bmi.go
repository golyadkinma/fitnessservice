// bmi.go — калькулятор индекса массы тела по шкале ВОЗ.
package server

import "fmt"

// BMIResult — результат расчёта индекса массы тела.
type BMIResult struct {
	Value    float64 `json:"value"`    // ИМТ, кг/м²
	Category string  `json:"category"` // категория по шкале ВОЗ
	Advice   string  `json:"advice"`   // краткая рекомендация
}

// Categories ВОЗ: <16 дефицит, 16–18.5 недостаточный, 18.5–25 норма,
// 25–30 избыточная, 30–35 ожирение I, 35–40 ожирение II, >40 ожирение III.
const (
	bmiDeficit      = 16.0
	bmiUnderweight  = 18.5
	bmiNormal       = 25.0
	bmiOverweight   = 30.0
	bmiObesityI     = 35.0
	bmiObesityII    = 40.0
	minWeightKg     = 20.0
	maxWeightKg     = 400.0
	minHeightM      = 0.5
	maxHeightM      = 2.6
)

// CalcBMI вычисляет ИМТ и возвращает категорию по шкале ВОЗ.
// Масса задаётся в килограммах, рост — в метрах.
func CalcBMI(weightKg, heightM float64) (BMIResult, error) {
	if weightKg < minWeightKg || weightKg > maxWeightKg {
		return BMIResult{}, fmt.Errorf("масса должна быть от %.0f до %.0f кг", minWeightKg, maxWeightKg)
	}
	if heightM < minHeightM || heightM > maxHeightM {
		return BMIResult{}, fmt.Errorf("рост должен быть от %.1f до %.1f м", minHeightM, maxHeightM)
	}
	value := weightKg / (heightM * heightM)
	res := BMIResult{Value: round2(value)}
	switch {
	case value < bmiDeficit:
		res.Category = "Выраженный дефицит массы тела"
		res.Advice = "Необходима консультация врача, повышение калорийности рациона"
	case value < bmiUnderweight:
		res.Category = "Недостаточная масса тела"
		res.Advice = "Увеличьте калорийность рациона и добавьте силовые тренировки"
	case value < bmiNormal:
		res.Category = "Нормальная масса тела"
		res.Advice = "Поддерживайте текущий режим питания и активности"
	case value < bmiOverweight:
		res.Category = "Избыточная масса тела"
		res.Advice = "Добавьте кардионагрузку и следите за суточной калорийностью"
	case value < bmiObesityI:
		res.Category = "Ожирение I степени"
		res.Advice = "Рекомендуется план снижения массы и контроль питания"
	case value < bmiObesityII:
		res.Category = "Ожирение II степени"
		res.Advice = "Необходима консультация врача и программа снижения массы"
	default:
		res.Category = "Ожирение III степени"
		res.Advice = "Обратитесь к врачу для составления программы лечения"
	}
	return res, nil
}

// round2 округляет число до двух знаков после запятой.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
