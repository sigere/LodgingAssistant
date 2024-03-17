package main

import (
	"errors"
	"net/url"
	"strconv"
)

type Rooms int

func (r Rooms) ToString() (string, error) {
	switch r {
	case 1:
		return "one", nil
	case 2:
		return "two", nil
	case 3:
		return "three", nil
	case 4:
		return "four", nil
	}

	return "", errors.New("unsupported rooms number")
}

type SearchConfig struct {
	Id            uint    `gorm:"primaryKey"`
	DistrictId    int     `gorm:"not null"`
	Rooms         Rooms   `gorm:"not null"`
	MinArea       int     `gorm:"not null"`
	MaxArea       int     `gorm:"not null"`
	MinPrice      int     `gorm:"not null"`
	MaxPrice      int     `gorm:"not null"`
	Furniture     bool    `gorm:"not null"`
	MaxTotalPrice float64 `gorm:"not null"`
}

func (config SearchConfig) ApplyOnUrl(values *url.Values) {
	values.Add("search[district_id]", strconv.Itoa(config.DistrictId))

	if config.Furniture {
		values.Add("search[filter_enum_furniture][0]", "yes")
	}

	rooms, err := config.Rooms.ToString()
	if err != nil {
		panic(err)
	}

	values.Add("search[filter_enum_rooms][0]", rooms)
	values.Add("search[filter_float_m:from]", strconv.Itoa(config.MinArea))
	values.Add("search[filter_float_m:to]", strconv.Itoa(config.MaxArea))
	values.Add("search[filter_float_price:from]", strconv.Itoa(config.MinPrice))
	values.Add("search[filter_float_price:to]", strconv.Itoa(config.MaxPrice))
}
