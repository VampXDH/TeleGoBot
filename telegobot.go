package telegobot

import (
    "fmt"
    "net/http"
    "encoding/json"
)

const telegramAPI = "https://api.telegram.org/bot"

type Bot struct {
    Token string
}

type Update struct {
    UpdateID int `json:"update_id"`
    Message  struct {
        Text string `json:"text"`
        Chat struct {
            ID int64 `json:"id"`
        } `json:"chat"`
    } `json:"message"`
}

type Response struct {
    Ok     bool     `json:"ok"`
    Result []Update `json:"result"`
}

func NewBot(token string) *Bot {
    return &Bot{
        Token: token,
    }
}

func (b *Bot) GetUpdates() ([]Update, error) {
    resp, err := http.Get(fmt.Sprintf("%s%s/getUpdates", telegramAPI, b.Token))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response Response
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    return response.Result, nil
}

func (b *Bot) SendMessage(chatID int64, text string) error {
    url := fmt.Sprintf("%s%s/sendMessage?chat_id=%d&text=%s", telegramAPI, b.Token, chatID, text)
    _, err := http.Get(url)
    return err
}
