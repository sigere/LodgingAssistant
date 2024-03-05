package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	fmt.Printf("Running at %s\n", time.Now().Format(time.RFC3339))
	db := GetDB()

	districtId := 255
	if len(os.Args) > 1 {
		districtId, _ = strconv.Atoi(os.Args[1])
	}

	rooms := 0
	if len(os.Args) > 2 {
		rooms, _ = strconv.Atoi(os.Args[2])
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

	realMaxPrice := 0
	if len(os.Args) > 6 {
		realMaxPrice, _ = strconv.Atoi(os.Args[6])
	}

	client := http.Client{}
	i := 1
	for {
		fmt.Println("Fetching page " + strconv.Itoa(i))
		responseHtml, err := fetch(&client, i, districtId, rooms, minArea, minPrice, maxPrice)
		if err != nil {
			panic(err)
		}
		os.WriteFile("./search_result.html", responseHtml, os.FileMode.Perm(0644))

		jsonData, err := extractData(responseHtml)
		if err != nil {
			panic(err)
		}

		marshalled, err := json.MarshalIndent(jsonData, "", " ")
		if err != nil {
			panic(err)
		}
		os.WriteFile("./test.json", marshalled, os.FileMode.Perm(0644))

		for _, ad := range jsonData.Listing.Listing.Ads {
			flat := ad.toFlat()
			flat.LastChecked = time.Now()

			err := db.Where(Flat{ExtId: uint(ad.Id)}).FirstOrCreate(flat).Error
			if err != nil {
				panic(err)
			}

			flat.LastChecked = time.Now()
			db.Save(flat)
		}

		i++
		if i > jsonData.Listing.Listing.TotalPages || i > 10 {
			break
		}
	}

	var subscriptions []Subscription
	db.Where("is_active", true).Find(&subscriptions)

	var flats []Flat
	db.Where("notified = ? and total_price <= ?", false, realMaxPrice).
		FindInBatches(&flats, 10, func(tx *gorm.DB, batch int) error {
			for i, flat := range flats {
				for _, subscription := range subscriptions {
					fmt.Println("Notifying", subscription.Address, flat.ExtId)
					err := notify(&flat, mail.Address{Address: subscription.Address})
					if err != nil {
						fmt.Println(err)
					}
				}

				flats[i].Notified = true
			}
			db.Save(flats)

			return nil
		})
}

func extractData(html []byte) (*OriginalJson, error) {
	pattern := regexp.MustCompile(`__PRERENDERED_STATE__= "(.*)";`)
	match := pattern.FindStringSubmatch(string(html))
	jsonData := OriginalJson{}
	if len(match) > 1 {
		jsonString := match[1]
		jsonString, _ = strconv.Unquote(`"` + jsonString + `"`)

		re := regexp.MustCompile(`\\(.)`)
		jsonString = re.ReplaceAllString(jsonString, `\$1`)

		if err := json.Unmarshal([]byte(jsonString), &jsonData); err != nil {
			panic(err)
		}
	} else {
		return nil, errors.New("")
	}

	return &jsonData, nil
}

func fetch(
	client *http.Client,
	page int,
	districtId int,
	rooms int,
	minArea int,
	minPrice int,
	maxPrice int,
) ([]byte, error) {
	request, err := http.NewRequest("GET", os.Getenv("BASE_URL"), nil)
	if err != nil {
		panic(err)
	}

	/*
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
	*/

	q := request.URL.Query()
	q.Add("page", strconv.Itoa(page))
	q.Add("search[district_id]", strconv.Itoa(districtId))
	q.Add("search[filter_enum_furniture][0]", "yes")

	q.Add("search[filter_enum_rooms][0]", "three")
	q.Add("search[filter_enum_rooms][1]", "four")

	/*
		todo
		if stringRooms != "" {
		q.Add("search[filter_enum_rooms][0]", stringRooms)
		}
	*/

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
