// catalog.go — каталог продукции: доменные типы, справочник по умолчанию, поиск и фильтры.
package server

import (
	"sort"
	"strings"
)

// FilterProducts возвращает товары по строке поиска и категории.
// Пустая строка поиска или категория означает «без условия».
func FilterProducts(products []Product, query, category string) []Product {
	query = strings.ToLower(strings.TrimSpace(query))
	category = strings.ToLower(strings.TrimSpace(category))
	result := make([]Product, 0, len(products))
	for _, p := range products {
		if category != "" && !strings.EqualFold(p.Category, category) {
			continue
		}
		if query != "" &&
			!strings.Contains(strings.ToLower(p.Name), query) &&
			!strings.Contains(strings.ToLower(p.Desc), query) {
			continue
		}
		result = append(result, p)
	}
	// Сортировка: сначала наличие на складе, затем по убыванию рейтинга.
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].InStock != result[j].InStock {
			return result[i].InStock
		}
		return result[i].Rating > result[j].Rating
	})
	return result
}

// Categories возвращает список уникальных категорий каталога.
func Categories(products []Product) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, p := range products {
		if !seen[p.Category] {
			seen[p.Category] = true
			result = append(result, p.Category)
		}
	}
	sort.Strings(result)
	return result
}
