package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var BASE_URL string = "https://www.tesla.com/inventory/api/v4/inventory-results"

// Model goes on the end (eg: m3)
var CAR_LINK_BASE_URL string = "https://www.tesla.com/en_GB"

// Structs for saving cars to a file
type SavedCar struct {
	Vin    string  `json:"vin"`
	Price  float64 `json:"price"`
	Photos bool    `json:"photos"`
}

type Car struct {
	// Basic car details
	Model string `json:"Model"`
	Year  int    `json:"Year"`

	// Car registration / identifier details
	Vin                 string `json:"VIN"`
	RegistrationDetails struct {
		LicensePlateNumber    string `json:"LicensePlateNumber"`
		FirstRegistrationDate string `json:"firstRegistered"`
	} `json:"RegistrationDetails"`

	// Car Details (Odometer, Trim, Options)
	Odometer        int    `json:"Odometer"`
	OdometerType    string `json:"OdometerType"`
	Trim            string `json:"TrimName"`
	OptionCodeSpecs struct {
		C_OPTS struct {
			Options []struct {
				Code         string `json:"code"`
				Name         string `json:"name"`
				LongName     string `json:"long_name"`
				Description  string `json:"description"`
				LexiconGroup string `json:"lexiconGroup"`
			} `json:"options"`
		} `json:"C_OPTS"`
	} `json:"OptionCodeSpecs"`

	// Car Location (Tesla Centre)
	Location string `json:"VrlName"`

	// Car Price & Link
	Price float64 `json:"InventoryPrice"`

	// Car Photos
	Photos []struct {
		PhotoURL string `json:"imageUrl"`
	} `json:"VehiclePhotos"`

	// Comparison to saved cars
	Is_new                   bool
	Price_changed_since_last bool
	Price_change             float64
	Photos_added_since_last  bool
	Missing_since_last       bool
}

func (c Car) SaveData() SavedCar {
	return SavedCar{
		Vin:    c.Vin,
		Price:  c.Price,
		Photos: len(c.Photos) > 0,
	}
}

type TeslaQueryOptions struct {
	Year []int    `json:"Year"`
	TRIM []string `json:"TRIM"`
}

type TeslaQuery struct {
	Model     string            `json:"model"`
	Condition string            `json:"condition"`
	Options   TeslaQueryOptions `json:"options"`
	Arrangeby string            `json:"arrangeby"`
	Order     string            `json:"order"`
	Market    string            `json:"market"`
}

type TeslaInventoryQuery struct {
	Query TeslaQuery `json:"query"`
}

type TeslaResponse struct {
	Results []Car `json:"results"`
}

func GetTeslaInventory(model string, years []int, trims []string) ([]Car, error) {
	// Define the output slice
	var output []Car

	// Get Tesla Inventory
	this_query := TeslaInventoryQuery{
		Query: TeslaQuery{
			Model:     model,
			Condition: "used",
			Options: TeslaQueryOptions{
				Year: years,
				TRIM: trims,
			},
			Arrangeby: "Year",
			Order:     "desc",
			Market:    "GB",
		},
	}

	// Convert query to json string for API request
	this_query_json, err := json.Marshal(this_query)
	if err != nil {
		fmt.Println("Error: ", err)
		return output, err
	}
	// return string(this_query_json)

	// Make API request to Tesla
	tr := &http.Transport{
		ForceAttemptHTTP2: false,
		Protocols:         new(http.Protocols),
	}
	tr.Protocols.SetHTTP2(false)

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", BASE_URL, nil)

	// Add headers to request so we don't get filtered
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0")

	// Add query string built above
	q := req.URL.Query()
	q.Add("query", string(this_query_json))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: ", err)
		return output, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: ", err)
		return output, err
	}
	// Now we have the body of the response, unmarshal it to a struct
	var response TeslaResponse
	json.Unmarshal([]byte(body), &response)

	// Using the result, we can now build the Car objects
	// fmt.Println(result["results"])
	// fmt.Println(response.Results)

	return response.Results, nil

}

func LoadSavedCars(inventory []Car) []Car {
	// Load the saved cars data from file
	jsonFile, err := os.ReadFile("cars.json")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	var loaded_cars map[string]SavedCar
	err = json.Unmarshal(jsonFile, &loaded_cars)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	// Now we have the loaded cars, we can compare them to the inventory
	for i, car := range inventory {
		// Check if the car is new
		if _, ok := loaded_cars[car.Vin]; ok {
			inventory[i].Is_new = false
			// Check if the price has changed
			if car.Price != loaded_cars[car.Vin].Price {
				inventory[i].Price_changed_since_last = true
				inventory[i].Price_change = car.Price - loaded_cars[car.Vin].Price
			}
			// Check if photos have been added
			if len(car.Photos) > 0 && !loaded_cars[car.Vin].Photos {
				inventory[i].Photos_added_since_last = true
			}
		} else {
			inventory[i].Is_new = true
		}
	}
	// Check for missing cars
	for vin, saved_car := range loaded_cars {
		var found bool = false
		for _, car := range inventory {
			if car.Vin == vin {
				found = true
				break
			}
		}
		if !found {
			// Add the missing car to the inventory
			inventory = append(inventory, Car{
				Vin:                vin,
				Price:              saved_car.Price,
				Is_new:             false,
				Missing_since_last: true,
			})
		}
	}
	return inventory
}

func SaveCars(inventory []Car) {
	// Save the cars to the file
	var saved_cars = make(map[string]SavedCar)
	for _, car := range inventory {
		if car.Missing_since_last {
			continue
		}
		saved_cars[car.Vin] = car.SaveData()
	}
	saved_cars_json, err := json.MarshalIndent(saved_cars, "", "    ")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	err = os.WriteFile("cars.json", saved_cars_json, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
