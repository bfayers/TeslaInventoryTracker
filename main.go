package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/bfayers/TeslaInventoryTracker/utils"
	"golang.org/x/text/message"
)

var DISCORD_WEBHOOK_URL string = os.Getenv("DISCORD_WEBHOOK_URL")
var DISCORD_NEW_CAR_THREAD string = os.Getenv("DISCORD_NEW_CAR_THREAD")
var DISCORD_CHANGED_CAR_THREAD string = os.Getenv("DISCORD_CHANGED_CAR_THREAD")

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type EmbedImage struct {
	Url string `json:"url"`
}

type EmbedAuthor struct {
	Name string `json:"name"`
}

type Embed struct {
	Title     string       `json:"title,omitempty,omitzero"`
	Url       string       `json:"url,omitempty,omitzero"`
	Color     int          `json:"color,omitempty,omitzero"`
	Author    EmbedAuthor  `json:"author,omitempty,omitzero"`
	Image     EmbedImage   `json:"image,omitempty,omitzero"`
	Thumbnail EmbedImage   `json:"thumbnail,omitempty,omitzero"`
	Fields    []EmbedField `json:"fields,omitempty,omitzero"`
}

func sendToDiscord(car utils.Car, threadId string) error {
	// Send to discord

	// Create the embed json

	// Outer data structure
	var data struct {
		Embeds []Embed `json:"embeds"`
	}

	p := message.NewPrinter(message.MatchLanguage("en"))
	embed := Embed{
		Title: p.Sprintf("%#d %s %s - £%.2f", car.Year, car.Model, car.Trim, car.Price),
		Url:   fmt.Sprintf("%s/%s/order/%s", utils.CAR_LINK_BASE_URL, car.Model, car.Vin),
		Color: 5814783,
		Author: EmbedAuthor{
			Name: car.Location,
		},
	}
	// Add the embed to the data
	data.Embeds = append(data.Embeds, embed)

	// Add the fields
	// VIN Field
	data.Embeds[0].Fields = append(data.Embeds[0].Fields, EmbedField{
		Name:   "VIN",
		Value:  car.Vin,
		Inline: true,
	})
	// Plate
	data.Embeds[0].Fields = append(data.Embeds[0].Fields, EmbedField{
		Name:   "Plate",
		Value:  car.RegistrationDetails.LicensePlateNumber,
		Inline: true,
	})

	// Odometer
	data.Embeds[0].Fields = append(data.Embeds[0].Fields, EmbedField{
		Name:   "Odometer",
		Value:  p.Sprintf("%d %s", car.Odometer, car.OdometerType),
		Inline: true,
	})

	// Options Field
	var options = EmbedField{
		Name:   "Options",
		Value:  "",
		Inline: false,
	}

	for _, option := range car.OptionCodeSpecs.C_OPTS.Options {
		options.Value += fmt.Sprintf("%s\n", option.Name)
	}
	data.Embeds[0].Fields = append(data.Embeds[0].Fields, options)

	// If the car has photos, add them to the embed
	if len(car.Photos) > 0 {
		data.Embeds[0].Thumbnail = EmbedImage{
			Url: car.Photos[0].PhotoURL,
		}
		data.Embeds[0].Image = EmbedImage{
			Url: car.Photos[1].PhotoURL,
		}
		for _, photo := range car.Photos[2:5] {
			data.Embeds = append(data.Embeds, Embed{
				Url: fmt.Sprintf("%s/%s/%s", utils.CAR_LINK_BASE_URL, car.Model, car.Vin),
				Image: EmbedImage{
					Url: photo.PhotoURL,
				},
			})
		}
	}

	// If this is an update to an existing car, add fields explaining the changes
	if !car.Is_new {
		// Price Change
		if car.Price_changed_since_last {
			var symbol = ""
			if car.Price_change > 0 {
				symbol = "+"
			}
			data.Embeds[0].Fields = append(data.Embeds[0].Fields, EmbedField{
				Name:   "Price Change",
				Value:  p.Sprintf("%s%.2f", symbol, car.Price_change),
				Inline: true,
			})
		}
		// Photos Added
		if car.Photos_added_since_last {
			data.Embeds[0].Fields = append(data.Embeds[0].Fields, EmbedField{
				Name:   "Photos Added",
				Value:  "Yes",
				Inline: true,
			})
		}
	}

	json_marshaled, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// Send the data to the discord webhook
	client := &http.Client{}
	req, _ := http.NewRequest("POST", DISCORD_WEBHOOK_URL, bytes.NewBuffer(json_marshaled))

	// Set the headers
	req.Header.Add("Content-Type", "application/json")

	// Add query string for thread
	q := req.URL.Query()
	q.Add("thread_id", threadId)
	req.URL.RawQuery = q.Encode()

	_, err = client.Do(req)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	} else {
		fmt.Printf("Sent %s to Discord\n", car.Vin)
	}
	return nil

}

func main() {
	inventory, err := utils.GetTeslaInventory("m3", []int{2024, 2025}, []string{"LRRWD", "LRAWD"})
	if err != nil {
		fmt.Println("Error: ", err)
	}
	// Load the saved cars
	inventory = utils.LoadSavedCars(inventory)

	var any_errors_sending bool = false
	// Send to discord
	for _, car := range inventory {
		if car.Is_new {
			err = sendToDiscord(car, DISCORD_NEW_CAR_THREAD)
		} else if car.Price_changed_since_last || car.Photos_added_since_last {
			err = sendToDiscord(car, DISCORD_CHANGED_CAR_THREAD)
		}
		if err != nil {
			fmt.Println("Error: ", err)
			any_errors_sending = true
		}
	}
	// Save the cars (if there were no errors sending)
	if !any_errors_sending {
		utils.SaveCars(inventory)
	}
}
