package service

import (
	"math"
	"regexp"
	"strings"
)

// countWords считает общее количество слов в тексте.
// Для точности используется регулярное выражение, чтобы считать только слова (буквы и цифры),
// игнорируя лишние пробелы и знаки препинания.
func countWords(text string) int {
	// Регулярное выражение для поиска последовательностей букв и цифр (слов)
	re := regexp.MustCompile(`[\w\p{L}]+`)
	words := re.FindAllString(text, -1)
	return len(words)
}

// countKeywordMatches считает, сколько ключевых слов из списка присутствует в тексте.
func countKeywordMatches(textToAnalyse string, genreKeywords string) int {
	text := strings.ToLower(textToAnalyse)
	keywords := strings.Split(strings.ToLower(genreKeywords), ",")

	matchCount := 0
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		// Проверяем, что ключевое слово не пустое и содержится в тексте
		if kw != "" && strings.Contains(text, kw) {
			matchCount++
		}
	}
	return matchCount
}

// CalculateGenreProbability рассчитывает вероятность принадлежности текста к жанру
// по формуле: P(gi|D) = Count(Keywords_i) / T (общее количество слов)
// Возвращает вероятность в процентах (0-100)
func CalculateGenreProbability(textToAnalyse string, genreKeywords string) int {
	if textToAnalyse == "" || genreKeywords == "" {
		return 0
	}

	// 1. Считаем количество совпадений ключевых слов (Count(Keywords_i))
	matchCount := countKeywordMatches(textToAnalyse, genreKeywords)

	// 2. Считаем общее количество слов в тексте (T)
	totalWords := countWords(textToAnalyse)

	if totalWords == 0 {
		return 0 // Избегаем деления на ноль
	}

	// 3. Расчет вероятности: P(gi|D) = Count(Keywords_i) / T
	probability := float64(matchCount) / float64(totalWords)

	// Преобразование в процент и округление
	probabilityPercent := int(math.Round(probability * 100))

	// Вероятность не может быть выше 100%
	if probabilityPercent > 100 {
		return 100
	}

	return probabilityPercent
}
