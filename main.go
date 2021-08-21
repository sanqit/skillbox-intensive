package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bResponse struct {
    Symbol string `json:"symbol"`
    Price float64 `json:"price,string"`
}

type wallet map[string]float64

var db = map[int]wallet{}

func main() {
    bot, err := tgbotapi.NewBotAPI(getToken())
    if err != nil {
        log.Panic(err)
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil {
            continue
        }

        command := strings.Split(update.Message.Text, " ")
        userId := update.Message.From.ID
        switch command[0] {
        case "ADD", "SUB":
            if len(command) != 3 {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
                continue
            }

            userId := update.Message.From.ID
            currency := command[1]
            _, err := getPrice(currency,"USDT")
            if err != nil {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
                continue
            }

            money, err := strconv.ParseFloat(command[2], 64)
            if err != nil {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
                continue
            }

            if _, ok := db[userId]; !ok {
                db[userId] = make(wallet)
            }

            switch command[0] {
            case "ADD":
                db[userId][currency] += money
            case "SUB":
                db[userId][currency] -= money
            }
        case "DEL":
            if len(command) != 2 {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
                continue
            }

            delete(db[userId], command[1])
        case "SHOW":
            fmt.Println(db[userId])
            resp := ""
            for key, value := range db[userId] {
                //targetCurrency := "USDT"
                targetCurrency := "RUB"
                usdPrice, err := getPrice(key, targetCurrency)
                if err != nil {
                    bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
                    continue
                }

                resp += fmt.Sprintf("%s: %s\n", key, formatPrice(value * usdPrice, targetCurrency))
            }

            if resp == "" {
                resp = "Кошелёк пуст"
            }

            bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resp))
        default:
            bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
        }
    }
}

func formatPrice(price float64, currency string) string {
    switch currency {
    case "RUB":
        return fmt.Sprintf("%.2f ₽", price)
    case "USDT":
        return fmt.Sprintf("$ %.2f", price)
    default:
        return fmt.Sprintf("%.2f", price)
    }
}

func getPrice(from string, to string) (float64, error) {
    url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s%s", from, to)
    resp, err := http.Get(url)
    if err != nil {
        return 0, err
    }

    var bRes bResponse
    err = json.NewDecoder(resp.Body).Decode(&bRes)
    if err != nil {
        return 0, err
    }

    if bRes.Symbol == "" {
        return 0, errors.New("Неверная валюта")
    }

    return bRes.Price, nil
}

func getToken() string {
    return "MyAwesomeBotToken"
}
