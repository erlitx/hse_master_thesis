package usecase

// Складывает два числа (демо)
func (u *UseCase) Add(a, b float64) float64 {
	return a + b
}

// Возвращает переданный текст (демо)
func (u *UseCase) Echo(text string) string {
	return text
}
