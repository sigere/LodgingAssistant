package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Ad struct {
	Id          int       `json:"id"`
	IsActive    bool      `json:"isActive"`
	CreatedTime time.Time `json:"createdTime"`
	PushupTime  time.Time `json:"pushupTime"`
	Description string    `json:"description"`
	Title       string    `json:"title"`
	Url         string    `json:"url"`
	Photos      []string  `json:"photos"`
	Location    struct {
		CityId               int    `json:"cityId"`
		CityNormalizedName   string `json:"cityNormalizedName"`
		DistrictId           int    `json:"districtId"`
		DistrictName         string `json:"districtName"`
		RegionId             int    `json:"regionId"`
		RegionNormalizedName string `json:"regionNormalizedName"`
	} `json:"location"`

	Map struct {
		Lat    float64 `json:"lat"`
		Lon    float64 `json:"lon"`
		Radius int     `json:"radius"`
	} `json:"map"`

	Params []struct {
		Key             string `json:"key"`
		Name            string `json:"name"`
		NormalizedValue string `json:"normalizedValue"`
		Value           string `json:"value"`
	} `json:"params"`

	Price struct {
		RegularPrice struct {
			Value int `json:"value"`
		} `json:"regularPrice"`
	} `json:"price"`
}

type OriginalJson struct {
	Listing struct {
		Listing struct {
			Ads        []Ad `json:"ads"`
			TotalPages int  `json:"totalPages"`
		} `json:"listing"`
	} `json:"listing"`
}

func (ad *Ad) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	*ad = Ad{}
	err := json.Unmarshal(bytes, ad)
	return err
}

func (ad Ad) Value() (driver.Value, error) {
	return json.Marshal(ad)
}

func (ad *Ad) toFlat() *Flat {
	flat := Flat{
		ExtId:      uint(ad.Id),
		IsActive:   ad.IsActive,
		TotalPrice: ad.getRent() + float64(ad.Price.RegularPrice.Value),
		CreatedAt:  ad.CreatedTime,
		PushedUpAt: sql.NullTime{
			Time:  ad.PushupTime,
			Valid: !ad.PushupTime.IsZero(),
		},
		Url:          ad.Url,
		Title:        ad.Title,
		Description:  ad.Description,
		Photos:       ad.Photos,
		OriginalJson: *ad,
	}

	return &flat
}

func (ad *Ad) getRent() float64 {
	for _, param := range ad.Params {
		if param.Key == "rent" {
			val, err := strconv.ParseFloat(param.NormalizedValue, 64)
			if err == nil {
				return val
			}
		}
	}

	return 0.0
}
