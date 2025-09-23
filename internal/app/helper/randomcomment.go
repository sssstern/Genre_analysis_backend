package helper

import "math/rand"

func GetRandomPhrase() string {
	phrases := []string{
		"срочно к проверке",
		"приоритет",
		"требует подтверждения",
		"вызывает сомнения",
		"",
		"интересный случай",
	}
	return phrases[rand.Intn(len(phrases))]
}
