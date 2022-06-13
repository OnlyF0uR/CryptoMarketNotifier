package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Lookup struct {
	USD float64 `json:"USD"`
}

func PriceInUSD(sym string) float64 {
	url := "https://min-api.cryptocompare.com/data/price?fsym=" + sym + "&tsyms=USD"
	req, ex := http.NewRequest(http.MethodGet, url, nil)
	if ex != nil {
		log.Fatal(ex)
	}

	res, ex := http.DefaultClient.Do(req)
	if ex != nil {
		log.Fatal(ex)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, ex := ioutil.ReadAll(res.Body)
	if ex != nil {
		log.Fatal(ex)
	}

	lookup := Lookup{}
	ex = json.Unmarshal(body, &lookup)
	if ex != nil {
		log.Fatal(ex)
	}

	return lookup.USD
}

var s *discordgo.Session

func main() {
	ex := godotenv.Load(".env")
	if ex != nil {
		log.Fatalf("Error occured while instatiating the bot: %v", ex)
	}

	s, ex := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if ex != nil {
		log.Fatalf("Error occured while instatiating the bot: %v", ex)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{{
				Name: "the market go up and down",
				Type: discordgo.ActivityTypeWatching,
				URL:  "",
			}},
		})
	})

	currencies := [4]string{"BTC", "ETH", "LTC", "XMR"}
	lastPrices := make(map[string]float64)

	fmt.Println("===================\nStarting prices:")

	isFirstRun := true

	for {
		// Predefine a message
		var msg string = ""

		// Loop through all the currencies we have
		for _, c := range currencies {
			// Get the price in usd
			pr := PriceInUSD(c)

			if isFirstRun {
				// If this is the first time just print the initial values
				fmt.Printf(" - %s: %.2f\n", c, pr)
			} else {
				// We calculate the difference
				diff := (pr - lastPrices[c]) / lastPrices[c]
				// fmt.Printf("%s: %.2f (%.2f%%)\n", c, pr, diff*100)

				// We add it onto the message
				msg += fmt.Sprintf("**%s:** %.2f (%.2f%%)\n", c, pr, diff*100)
			}

			// Save the last price
			lastPrices[c] = pr
		}

		if isFirstRun {
			isFirstRun = false
		} else {
			// Try sending the message
			_, ex := s.ChannelMessageSendComplex(os.Getenv("CHANNEL_ID"), &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Market Update",
						Description: msg,
						Color:       0x2f3136,
					},
				},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label: "BTC",
								Style: discordgo.LinkButton,
								URL:   "https://coinmarketcap.com/currencies/bitcoin/",
							},
							discordgo.Button{
								Label: "ETH",
								Style: discordgo.LinkButton,
								URL:   "https://coinmarketcap.com/currencies/ethereum/",
							},
							discordgo.Button{
								Label: "LTC",
								Style: discordgo.LinkButton,
								URL:   "https://coinmarketcap.com/currencies/litecoin/",
							},
							discordgo.Button{
								Label: "XMR",
								Style: discordgo.LinkButton,
								URL:   "https://coinmarketcap.com/currencies/monero/",
							},
						},
					},
				},
			})

			if ex != nil {
				log.Fatalf("Failed to send message: %v", ex)
			}
		}

		time.Sleep(time.Hour)
		// time.Sleep(15 * time.Second)
	}
}
