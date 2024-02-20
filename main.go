package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

const BASE_URL = "https://www.olx.pl/nieruchomosci/mieszkania/wynajem/krakow/"

type Ad struct {
	Id          int    `json:"id"`
	IsActive    bool   `json:"isActive"`
	CreatedTime string `json:"createdTime"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Url         string `json:"url"`
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

var paramTypes map[string]string
var districts map[int]string

func main() {
	initialize()

	districtId := 255
	if len(os.Args) > 1 {
		districtId, _ = strconv.Atoi(os.Args[1])
	}

	rooms := 0
	if len(os.Args) > 2 {
		rooms, _ = strconv.Atoi(os.Args[1])
	}

	minArea := 0
	if len(os.Args) > 3 {
		minArea, _ = strconv.Atoi(os.Args[3])
	}

	minPrice := 0
	if len(os.Args) > 4 {
		minPrice, _ = strconv.Atoi(os.Args[4])
	}

	maxPrice := 0
	if len(os.Args) > 5 {
		maxPrice, _ = strconv.Atoi(os.Args[5])
	}

	client := http.Client{}
	responseHtml, err := search(&client, districtId, rooms, minArea, minPrice, maxPrice)
	if err != nil {
		panic(err)
	}

	os.WriteFile("./search_result.html", responseHtml, os.FileMode.Perm(0644))

	pattern := regexp.MustCompile(`__PRERENDERED_STATE__= "(.*)";`)
	match := pattern.FindStringSubmatch(string(responseHtml))
	jsonData := OriginalJson{}
	//jsonData := map[string]interface{}{}
	if len(match) > 1 {
		jsonString := match[1]
		jsonString, _ = strconv.Unquote(`"` + jsonString + `"`)

		re := regexp.MustCompile(`\\(.)`)
		jsonString = re.ReplaceAllString(jsonString, `\$1`)

		if err := json.Unmarshal([]byte(jsonString), &jsonData); err != nil {
			panic(err)
		}

		//for _, ad := range jsonData.Listing.Listing.Ads {
		//	fmt.Println(ad.Description)
		//	fmt.Println("--------------")
		//}

		file, _ := json.MarshalIndent(jsonData, "", " ")
		_ = os.WriteFile("test.json", file, 0644)
	} else {
		fmt.Println("No match found.")
	}

	//fmt.Println("Found " + strconv.Itoa(len(jsonData.Listing.Listing.Ads)) + " adverts")
}

func search(client *http.Client, districtId int, rooms int, minArea int, minPrice int, maxPrice int) ([]byte, error) {
	request, err := http.NewRequest("GET", BASE_URL, nil)
	if err != nil {
		panic(err)
	}

	var stringRooms = ""
	switch rooms {
	case 1:
		stringRooms = "one"
	case 2:
		stringRooms = "two"
	case 3:
		stringRooms = "three"
	case 4:
		stringRooms = "four"
	}

	q := request.URL.Query()
	q.Add("search[district_id]", strconv.Itoa(districtId))
	q.Add("search[filter_enum_furniture][0]", "yes")

	if stringRooms != "" {
		q.Add("search[filter_enum_rooms][0]", stringRooms)
	}

	if minArea > 0 {
		q.Add("search[filter_float_m:from]", strconv.Itoa(minArea))
	}

	if minPrice > 0 {
		q.Add("search[filter_float_price:from]", strconv.Itoa(minPrice))
	}

	if maxPrice > 0 {
		q.Add("search[filter_float_price:to]", strconv.Itoa(maxPrice))
	}

	request.URL.RawQuery = q.Encode()
	fmt.Println(request.URL.String())
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func initialize() {
	paramTypes = map[string]string{
		"PARAM_FLOOR":      "floor_select",
		"PARAM_FURNITURE":  "furniture",
		"PARAM_BUILT_TYPE": "builttype",
		"PARAM_AREA":       "m",
		"PARAM_ROOMS":      "rooms",
		"PARAM_RENT":       "rent",
	}

	districts = map[int]string{
		255: "Krowodrza",
		259: "Bronowice",
	}
}
