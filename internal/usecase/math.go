package usecase

func (u *UseCase) Add(a, b float64) float64 {
	return a + b
}

func (u *UseCase) Echo(text string) string {
	return text
}
