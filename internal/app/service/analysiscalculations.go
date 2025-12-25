package service

import (
	"math"
	"regexp"
	"strings"
)

func countWords(text string) int {
	re := regexp.MustCompile(`[\w\p{L}]+`)
	words := re.FindAllString(text, -1)
	return len(words)
}

func countKeywordMatches(textToAnalyse string, genreKeywords string) int {
	text := strings.ToLower(textToAnalyse)
	keywords := strings.Split(strings.ToLower(genreKeywords), ",")

	matchCount := 0
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" && strings.Contains(text, kw) {
			matchCount++
		}
	}
	return matchCount
}

func CalculateGenreProbability(textToAnalyse string, genreKeywords string) int {
	if textToAnalyse == "" || genreKeywords == "" {
		return 0
	}

	matchCount := countKeywordMatches(textToAnalyse, genreKeywords)
	totalWords := countWords(textToAnalyse)

	if totalWords == 0 {
		return 0
	}

	probability := float64(matchCount) / float64(totalWords)
	probabilityPercent := int(math.Round(probability * 100))

	if probabilityPercent > 100 {
		return 100
	}

	return probabilityPercent
}
