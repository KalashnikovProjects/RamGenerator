package entities

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type IdResponse struct {
	Id int
}

type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`           // Максимум 24 символа
	Password     string `json:"password,omitempty"` // В базе данных PasswordHash, Password только в запросах регистрации / входа, put user, patch user
	PasswordHash string `json:"-"`

	DailyRamGenerationTime int `json:"daily_ram_generation_time"` // Время генерации первого барана за день, UNIX формат, изменяется только при ws/create-ram
	RamsGeneratedLastDay   int `json:"rams_generated_last_day"`   // Изменяется только при ws/generate-ram
	ClickersBlockedUntil   int `json:"-"`                         // Нельзя иметь 2 ws/generate-ram на аккаунт одновременно

	AvatarRamId int  `json:"avatar_ram_id"`
	AvatarBox   *Box `json:"avatar_box"` // 4 координаты, обрезающие аватар, координаты от 0 до 1
}

func (u *User) CalculateRamsGeneratedLastDay(timeBetweenDaily int) int {
	if u.DailyRamGenerationTime+timeBetweenDaily*60 < int(time.Now().Unix()) {
		return 0
	}
	return u.RamsGeneratedLastDay
}

type Ram struct {
	Id          int    `json:"id"`
	Taps        int    `json:"taps"`        // Изменяется только через ws/clicker
	Description string `json:"description"` // Оно же промпт для нейросети
	ImageUrl    string `json:"image_url"`
	UserId      int    `json:"user_id"` // Изменяется только через трейды
}

type Box [][]float64

func (b *Box) String() string {
	return fmt.Sprintf("((%g,%g),(%g,%g))", (*b)[0][0], (*b)[0][1], (*b)[1][0], (*b)[1][1])
}

func (b *Box) JsonString() string {
	return fmt.Sprintf("[[%g,%g],[%g,%g]])", (*b)[0][0], (*b)[0][1], (*b)[1][0], (*b)[1][1])
}

func NewBox(s string) (*Box, error) {
	s = strings.Replace(s, " ", "", -1)[1 : len(s)-1]
	splitS := strings.Split(s, "),(")[:2]

	strNums := append(strings.Split(splitS[0], ",")[:2], strings.Split(splitS[1], ",")[:2]...)
	numNums := make([]float64, 0, 4)
	for _, n := range strNums {
		floatN, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return &Box{}, err
		}
		numNums = append(numNums, floatN)
	}
	return &Box{{numNums[0], numNums[1]}, {numNums[2], numNums[3]}}, nil
}
