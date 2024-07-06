package entities

import (
	"fmt"
	"strconv"
	"strings"
)

type IdResponse struct {
	Id int
}

type User struct {
	Id               int    `json:"-"`
	Username         string `json:"username"`           // Максимум 24 символа
	Password         string `json:"password,omitempty"` // Для базы данных есть только PasswordHash, Password есть только для запросов регистрации / входа, put, patch
	PasswordHash     string `json:"-"`
	LastRamGenerated int    `json:"last_ram_generated"` // UNIX формат, изменяется только при POST нового барана
	AvatarUrl        string `json:"avatar_url"`
	AvatarBox        *Box   `json:"avatar_box"`
}

type Ram struct {
	Id          int    `json:"id"`
	Description string `json:"description"` // Оно же промпт для нейросети
	ImageUrl    string `json:"image_url"`
	UserId      int    `json:"-"` // Редактировать можно только userId
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
