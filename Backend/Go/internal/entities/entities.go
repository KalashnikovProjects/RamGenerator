package entities

type User struct {
	Id               int
	Username         string // Максимум 24 символа
	Password         string // Для базы данных есть только PasswordHash, Password есть только для запросов регистрации / входа
	PasswordHash     string
	LastRamGenerated int // UNIX формат
	AvatarRamId      int
	AvatarBox        [2][2]float32
}

type Ram struct {
	Id          int
	Description string // Оно же промпт для нейросети
	ImageUrl    string
	UserId      int
}
