package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bnResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

type wallet map[string]float64

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("tokenTelegaSkillbox"))
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		command := strings.Split(update.Message.Text, " ")

		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верная команда"))
				break
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}
			db[update.Message.Chat.ID][command[1]] += amount
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(db[update.Message.Chat.ID][command[1]])))
		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верная команда"))
				break
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				continue
			}
			db[update.Message.Chat.ID][command[1]] -= amount
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(db[update.Message.Chat.ID][command[1]])))
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верная команда"))
				break
			}
			delete(db[update.Message.Chat.ID], command[1])
		case "SHOW":
			msg := ""
			var sum float64 = 0
			for key, value := range db[update.Message.Chat.ID] {
				price, err := getPrice(key)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				fmt.Println(price)
				sum += value * price
				msg += fmt.Sprintf("%s: %f [%.2f]\n", key, value, value*price)
			}
			msg += fmt.Sprintf("Total: %.2f\n", sum)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не найдена"))

		}

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, command[0])
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)
	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sRUB", symbol))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		err = fmt.Errorf("не верный символ")
		return
	}

	price = jsonResp.Price
	return
}
