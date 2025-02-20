package main

import (
	"os"
	"log"
	"github.com/joho/godotenv"
	tb "github.com/go-telegram-bot-api/telegram-bot-api"
)	

func main() {
	token, err:= GetEnv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("No token")
}
	if err != nil{
		log.Panic(err)
}

	bot, err := tb.NewBotAPI(token)
	if err != nil{
		log.Panic(err)
}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u :=tb.NewUpdate(0)
	u.Timeout =60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil{
		log.Panic(err)
}

	for update:= range updates{
	if update.CallbackQuery == nil{
	log.Println(update.Message.From.ID,update.Message.From.UserName, update.Message.Text)
}
	 if update.CallbackQuery != nil {
            callbackData := update.CallbackQuery.Data
            switch callbackData {
            default:
                log.Printf("%s Unknown command: %s", update.CallbackQuery.From.UserName,callbackData)
            }
        } else if update.Message != nil && update.Message.IsCommand() {
            command := update.Message.Command()
            switch command {
		case "start":
			sendStartKeyboard(update.Message, bot)
            default:
                log.Printf("Unknown command: %s", command)
		}

	}
}
}
func sendStartKeyboard(message *tb.Message, bot *tb.BotAPI){
	btn1:=tb.NewInlineKeyboardButtonData("Выбрать город", "chooseCity")
	btn2:=tb.NewInlineKeyboardButtonData("Выбрать концерт", "chooseConcert")
	kb := [][]tb.InlineKeyboardButton{{btn1},{ btn2}}
	markup := &tb.InlineKeyboardMarkup{
		InlineKeyboard: kb,
	}
	msg:=tb.NewMessage(message.Chat.ID, "Добро пожаловать!")
	msg.ReplyMarkup=markup
	if _, err:= bot.Send(msg); err!=nil{
	log.Panic(err)
	}
}

type getEnvError struct{}
func (e getEnvError)Error() string{
	return "Не удалось загрузить .env"
}
func GetEnv(key string)(string,error){
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal("не удалось загрузить .env", err)
		return "", getEnvError{}
	}
	value:=os.Getenv(key)
	return value, nil
}
