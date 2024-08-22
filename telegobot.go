package telegobot

import (
    "fmt"
    "net/http"
    "encoding/json"
    "os"
)

const telegramAPI = "https://api.telegram.org/bot"

type Bot struct {
    Token    string
    Commands map[string]func(chatID int64, bot *Bot) string
}

type Update struct {
    UpdateID int `json:"update_id"`
    Message  struct {
        Text     string `json:"text"`
        Chat     struct {
            ID int64 `json:"id"`
        } `json:"chat"`
        Document struct {
            FileID   string `json:"file_id"`
            FileName string `json:"file_name"`
        } `json:"document"`
    } `json:"message"`
}

type Response struct {
    Ok     bool     `json:"ok"`
    Result []Update `json:"result"`
}

type FileResponse struct {
    Ok     bool `json:"ok"`
    Result struct {
        FilePath string `json:"file_path"`
    } `json:"result"`
}

func NewBot(token string) *Bot {
    bot := &Bot{
        Token:    token,
        Commands: make(map[string]func(chatID int64, bot *Bot) string),
    }

    return bot
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

func (b *Bot) SendFile(chatID int64, filePath string) error {
    url := fmt.Sprintf("%s%s/sendDocument", telegramAPI, b.Token)
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Placeholder for actual file upload logic
    // You would use multipart form here to upload the file to Telegram server

    return nil
}

func (b *Bot) GetFileURL(fileID string) (string, error) {
    resp, err := http.Get(fmt.Sprintf("%s%s/getFile?file_id=%s", telegramAPI, b.Token, fileID))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var fileResp FileResponse
    if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
        return "", err
    }

    if !fileResp.Ok {
        return "", fmt.Errorf("failed to get file URL")
    }

    return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token, fileResp.Result.FilePath), nil
}

func (b *Bot) RegisterCommand(command string, action func(chatID int64, bot *Bot) string) {
    b.Commands[command] = action
}

func (b *Bot) HandleUpdate(update Update) {
    chatID := update.Message.Chat.ID
    text := update.Message.Text

    // Handle commands
    if strings.HasPrefix(text, "/") {
        command := strings.Fields(text)[0]
        if action, exists := b.Commands[command]; exists {
            response := action(chatID, b)
            b.SendMessage(chatID, response)
        } else {
            b.SendMessage(chatID, "Unknown command. Use /help to see available commands.")
        }
    }

    // Handle file uploads
    if update.Message.Document.FileID != "" {
        b.SendMessage(chatID, "Processing your file...")
        fileURL, err := b.GetFileURL(update.Message.Document.FileID)
        if err != nil {
            b.SendMessage(chatID, "Failed to retrieve file.")
            return
        }

        // Pass file URL to main.go for proxy checking
        activeProxies, err := CheckProxies(fileURL)
        if err != nil {
            b.SendMessage(chatID, "Failed to check proxies.")
            return
        }

        resultFileName := "active_proxies.txt"
        err = SaveProxies(resultFileName, activeProxies)
        if err != nil {
            b.SendMessage(chatID, "Failed to save active proxies.")
            return
        }

        b.SendFile(chatID, resultFileName)
    }
}
