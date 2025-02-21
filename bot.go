package main

import (
	"io"

	"bytes"
	"net/http"
	"os"
	"log"
	"github.com/joho/godotenv"
	"github.com/mattn/go-sqlite3"
	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	updates := bot.GetUpdatesChan(u)


	for update:= range updates{
	if update.CallbackQuery == nil{
	log.Println(update.Message.From.UserName, update.Message.Text)
}
	 if update.CallbackQuery != nil {
		handleCallbacks(bot, &update)
        } else if update.Message != nil && update.Message.IsCommand() {
		handleCommands(bot, &update)
		}

	}
}



func handleCommands(bot *tb.BotAPI, update *tb.Update){
command := update.Message.Command()
            switch command {
		case "start":
			sendStartKeyboard(update.Message, bot, "Добро пожаловать!")
            default:
                log.Printf("Unknown command: %s", command)
		}
}



func handleCallbacks(bot *tb.BotAPI, update *tb.Update){
	clientID:= update.CallbackQuery.Message.Chat.ID
	messageID:=update.CallbackQuery.Message.MessageID
	clientUserName:=update.CallbackQuery.From.UserName
	callbackData:=update.CallbackQuery.Data
            switch callbackData {
		case "chooseCity":
			newButtons :=[][]tb.InlineKeyboardButton{
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Омск", "Omsk"),
				},
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Другое","other"),
				},
			}
			markup:=&tb.InlineKeyboardMarkup{
				InlineKeyboard: newButtons,
			}
			newmsg :=tb.NewEditMessageText(
				clientID,
				messageID,
				"Выбери доступный город:",
			)
			newmsg.ReplyMarkup=markup
			log.Println(clientUserName, callbackData)
			if _,err:=bot.Send(newmsg); err!=nil{
				log.Panic(err)
			}
		case "Omsk":
			newButtons :=[][]tb.InlineKeyboardButton{
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать город", "chooseCity"),
				},
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать концерт","chooseConcert"),
				},
			}
			markup:=&tb.InlineKeyboardMarkup{
				InlineKeyboard: newButtons,
			}
			newmsg :=tb.NewEditMessageText(
				clientID,
				messageID,
				"Вы выбрали Омск",
			)
			newmsg.ReplyMarkup=markup
			log.Println(clientUserName, callbackData)
			if _,err:=bot.Send(newmsg); err!=nil{
				log.Panic(err)
			}
		case "other":
			newButtons :=[][]tb.InlineKeyboardButton{
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать город", "chooseCity"),
				},
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать концерт","chooseConcert"),
				},
			}
			markup:=&tb.InlineKeyboardMarkup{
				InlineKeyboard: newButtons,
			}
			newmsg :=tb.NewEditMessageText(
				clientID,
				messageID,
				"К сожалению в вашем городе нету мероприятий, подключенных к нам",
			)
			newmsg.ReplyMarkup=markup
			log.Println(clientUserName, callbackData)
			if _,err:=bot.Send(newmsg); err!=nil{
				log.Panic(err)
			}

		case "chooseConcert":
			newButtons :=[][]tb.InlineKeyboardButton{
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Кьюар бар 29 февраля", "QRbar29"),
				},
			}
			markup:=&tb.InlineKeyboardMarkup{
				InlineKeyboard: newButtons,
			}
			newmsg :=tb.NewEditMessageText(
				clientID,
				messageID,
				"Выберите концерт:",
			)
			newmsg.ReplyMarkup=markup
			log.Println(clientUserName, callbackData)
			if _,err:=bot.Send(newmsg); err!=nil{
				log.Panic(err)
			}
		case "QRbar29":
			url := "http://localhost:8080/qr?name=" + clientUserName
			resp, err:= http.Get(url)
			if err!= nil{
			log.Panic(err)
			}
			defer resp.Body.Close()
			img := new(bytes.Buffer)
			_, err = io.Copy(img,resp.Body)
			if err !=nil{
			log.Panic(err)
			}
			msg:= tb.FileBytes{"image.jpg", img.Bytes()}
			sendimage:=tb.NewPhoto(clientID,msg)
			bot.Send(sendimage)
								
            default:
                log.Printf("%s Unknown command: %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
            }
}



func sendStartKeyboard(message *tb.Message, bot *tb.BotAPI, startmsg string){
	btn1:=tb.NewInlineKeyboardButtonData("Выбрать город", "chooseCity")
	btn2:=tb.NewInlineKeyboardButtonData("Выбрать концерт", "chooseConcert")
	kb := [][]tb.InlineKeyboardButton{{btn1},{ btn2}}
	markup := &tb.InlineKeyboardMarkup{
		InlineKeyboard: kb,
	}
	msg:=tb.NewMessage(message.Chat.ID, startmsg)
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
