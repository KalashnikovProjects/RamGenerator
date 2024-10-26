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
	Username     string `json:"username"`           // Max 24 symbols
	Password     string `json:"password,omitempty"` // In database saving only PasswordHash, Password using only in login, registration and put user, patch user request json's
	PasswordHash string `json:"-"`

	DailyRamGenerationTime int `json:"daily_ram_generation_time"` // Time of generating first daily ram, UNIX format, updating only with ws/generate-ram
	RamsGeneratedLastDay   int `json:"rams_generated_last_day"`   // Updating only with ws/generate-ram
	ClickersBlockedUntil   int `json:"-"`                         // User cant run 2 ws/generate-ram per account parallel

	AvatarRamId int    `json:"avatar_ram_id"`
	AvatarBox   *Box   `json:"avatar_box,omitempty"` // coordinates in format [[Left, Up], [Right, Bottom]] from 0 to 1
	AvatarUrl   string `json:"avatar_url,omitempty"`
}

func (u *User) CalculateRamsGeneratedLastDay(timeBetweenDaily int) int {
	if u.DailyRamGenerationTime+timeBetweenDaily*60 < int(time.Now().Unix()) {
		return 0
	}
	return u.RamsGeneratedLastDay
}

type Ram struct {
	Id          int    `json:"id"`
	Taps        int    `json:"taps"`        // Changing only with ws/clicker
	Description string `json:"description"` // Generating after ram image by ai
	ImageUrl    string `json:"image_url"`
	User        *User  `json:"user,omitempty"` // Only returning, not stores in db or input
	UserId      int    `json:"user_id"`
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
