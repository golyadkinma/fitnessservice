// types.go — доменные типы приложения и стартовое наполнение справочников.
// Все структуры собраны в одном файле, чтобы зависимости между подсистемами
// оставались явными, а промежуточные коммиты оставались компилируемыми.
package server

// Формат даты во всех журналах приложения.
const dateLayout = "2006-01-02"

// Product — товар спортивного питания в каталоге.
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"` // протеин, витамины, изотоник, снек, экипировка
	Price    float64 `json:"price"`    // цена, руб.
	Rating   float64 `json:"rating"`   // рейтинг 0..5
	InStock  bool    `json:"inStock"`
	Desc     string  `json:"desc"`
}

// StepDay — запись журнала шагов за один день.
type StepDay struct {
	Date  string `json:"date"`  // дата в формате ГГГГ-ММ-ДД
	Steps int    `json:"steps"` // количество шагов за день
}

// Reminder — напоминание с указанием дней недели и времени.
type Reminder struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Time   string `json:"time"`  // ЧЧ:ММ
	Days   []int  `json:"days"`  // 0 = воскресенье, 1 = понедельник, ..., 6 = суббота
	Active bool   `json:"active"`
}

// Challenge — челлендж с целевым количеством шагов на период.
type Challenge struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Target    int    `json:"target"`    // суммарная цель, шагов
	StartDate string `json:"startDate"` // ГГГГ-ММ-ДД
	EndDate   string `json:"endDate"`   // ГГГГ-ММ-ДД
	Desc      string `json:"desc"`
}

// ChallengePace — расчёт дневного темпа для достижения цели.
type ChallengePace struct {
	DaysLeft  int     `json:"daysLeft"`  // осталось дней
	Remaining int     `json:"remaining"` // осталось шагов до цели
	PerDay    int     `json:"perDay"`    // необходимый темп, шагов в день
	DonePct   float64 `json:"donePct"`   // процент выполнения
}

// UserProgress — очки, уровень и достижения пользователя.
type UserProgress struct {
	Points        int      `json:"points"`        // всего очков
	Level         int      `json:"level"`         // текущий уровень
	LevelProgress int      `json:"levelProgress"` // прогресс внутри уровня, 0..100
	Achievements  []string `json:"achievements"`  // коды полученных достижений
}

// Achievement — достижение и условие его получения.
type Achievement struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Points   int    `json:"points"`   // очки за получение
	MinSteps int    `json:"minSteps"` // дневная норма шагов для получения
}

// ChatMessage — сообщение общего чата.
type ChatMessage struct {
	ID   int    `json:"id"`
	User string `json:"user"` // имя автора
	Text string `json:"text"`
	At   string `json:"at"` // время в формате RFC3339
}

// Profile — анкетные данные пользователя для расчётов.
type Profile struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"` // масса, кг
	Height float64 `json:"height"` // рост, м
	Goal   int     `json:"goal"`   // дневная норма шагов
	Birth  string  `json:"birth"`  // дата рождения, ГГГГ-ММ-ДД
}

// defaultProducts — стартовое наполнение каталога.
func defaultProducts() []Product {
	return []Product{
		{1, "Протеин Whey Isolate 900 г", "протеин", 3490, 4.8, true, "Изолят сывороточного белка, 90% белка, без лактозы"},
		{2, "Протеин Casein Night 700 г", "протеин", 2790, 4.6, true, "Казеиновый протеин для приёма на ночь"},
		{3, "BCAA 2:1:1 400 г", "аминокислоты", 1890, 4.4, true, "Ветвление аминокислоты с глутамином"},
		{4, "Мультивитамины Daily 60 капс", "витамины", 990, 4.5, true, "Дневная норма витаминов и минералов"},
		{5, "Витамин D3 2000 МЕ", "витамины", 590, 4.7, true, "Поддержка иммунитета и костной ткани"},
		{6, "Изотоник Citrus 500 г", "изотоник", 890, 4.3, true, "Электролиты и углеводы для тренировок"},
		{7, "Изотоник Berry 500 г", "изотоник", 890, 4.2, false, "Электролиты и углеводы для тренировок"},
		{8, "Протеиновый батончик Choco", "снек", 149, 4.1, true, "20 г белка, без сахара"},
		{9, "Ореховая паста Protein 300 г", "снек", 690, 4.5, true, "Арахисовая паста с добавленным белком"},
		{10, "Шейкер SportMix 700 мл", "экипировка", 790, 4.0, true, "Шейкер с отсеком для порошка"},
		{11, "Коврик для йоги Pro", "экипировка", 2190, 4.6, true, "Нескользящий коврик 6 мм с чехлом"},
		{12, "Гантели наборные 2x10 кг", "экипировка", 5490, 4.8, false, "Обрезиненные диски, стальной гриф"},
	}
}

// defaultChallenges — стартовый набор челленджей.
func defaultChallenges() []Challenge {
	return []Challenge{
		{"week-56k", "56 000 шагов за неделю", 56000, "2026-10-26", "2026-11-01", "Держите норму 8 000 шагов в день — и цель достигнута"},
		{"no-elevator", "Неделя без лифта", 35000, "2026-10-26", "2026-11-01", "Ступеньки вместо лифта: 5 000 шагов в день"},
		{"november-100k", "100 000 шагов за ноябрь", 100000, "2026-11-01", "2026-11-30", "Месячный марафон для терпеливых"},
	}
}
