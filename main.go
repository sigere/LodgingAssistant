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
	if len(os.Args) != 2 {
		fmt.Println("Exactly one argument required [notify|collect]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "notify":
		notify()
	case "collect":
		collect()
	default:
		fmt.Println("Invalid argument. Use one of [notify|collect]")
		os.Exit(1)
	}
}

func collect() {
	fmt.Printf("Collecting at %s\n", time.Now().Format(time.RFC3339))
	db := GetDB()

	var configs []SearchConfig
	db.Find(&configs)
	client := http.Client{}

	for _, config := range configs {
		i := 1
		for {
			fmt.Println("Fetching page " + strconv.Itoa(i))
			responseHtml, err := fetch(&client, i, config)
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

				if flat.TotalPrice > config.MaxTotalPrice {
					continue
				}

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
	}
}

func notify() {
	fmt.Printf("Notifying at %s\n", time.Now().Format(time.RFC3339))
	db := GetDB()

	var subscriptions []Subscription
	db.Where("is_active", true).Find(&subscriptions)

	var flats []Flat
	db.Where("notified = ?", false).
		FindInBatches(&flats, 10, func(tx *gorm.DB, batch int) error {
			for i, flat := range flats {
				for _, subscription := range subscriptions {
					fmt.Println("Notifying", subscription.Address, flat.ExtId)
					err := notifyOne(&flat, mail.Address{Address: subscription.Address})
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

func fetch(client *http.Client, page int, config SearchConfig) ([]byte, error) {
	request, err := http.NewRequest("GET", os.Getenv("BASE_URL"), nil)
	if err != nil {
		panic(err)
	}

	q := request.URL.Query()
	q.Add("page", strconv.Itoa(page))
	config.ApplyOnUrl(&q)

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
