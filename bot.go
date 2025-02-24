package main

import (
	"crypto/rand"
	"database/sql"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func main() {
	dburl, err := GetEnv("DB") // get environment DB
	if err != nil {
		log.Panic(err)
	}
	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	token, err := GetEnv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("No token")
	}
	if err != nil {
		log.Panic(err)
	}

	bot, err := tb.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tb.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery == nil {
			log.Println(update.Message.From.ID, update.Message.From.UserName, update.Message.Text)
		}
		if update.CallbackQuery != nil {
			handleCallbacks(db, bot, &update)
		} else if update.Message != nil && update.Message.IsCommand() {
			handleCommands(bot, db, &update)
		}

	}
}

func insertUser(db *sql.DB, login int64, root bool, admin bool, city string, password string) {
	sqlStmt := "INSERT INTO public.users (login, root, citycode, adm, password_hash) VALUES($1, $2, $3, $4, $5);"
	log.Println(sqlStmt, login, root, city, admin, password)
	_, err := db.Exec(sqlStmt, login, root, city, admin, password)
	if err != nil {
		log.Panic(err)
	}
}
func insertCity(db *sql.DB, cityCode string, cityName string) {
	sqlStmt := "INSERT INTO public.cities (city, cityname) VALUES ($1, $2)"
	log.Println(sqlStmt, cityCode, cityName)
	_, err := db.Exec(sqlStmt, cityCode, cityName)
	if err != nil {
		log.Panic(err)
	}
}

func generateTempPassword(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[num.Int64()]
	}
	return string(b)
}

func handleCommands(bot *tb.BotAPI, db *sql.DB, update *tb.Update) {
	command := update.Message.Command()
	clientLogin := update.Message.From.ID
	switch command {
	case "start":
		row := db.QueryRow("SELECT login FROM public.users WHERE login = $1", clientLogin)
		err := row.Scan(&clientLogin)
		if err != nil {
			if err == sql.ErrNoRows {
				password := (generateTempPassword(16))
				insertUser(db, update.Message.From.ID, false, false, "Omsk", password)
			} else {
				log.Panic(err)
			}
		}

		SendInlineKeyboard(bot, update,
			[][]tb.InlineKeyboardButton{
				{tb.NewInlineKeyboardButtonData("Выбрать город", "choose/cities/0")},
				{tb.NewInlineKeyboardButtonData("Выбрать концерт", "chooseConcert")}},
			"Добро пожаловать!",
		)
	case "root":
		if isRoot(bot, update, db, clientLogin, command) {
			SendInlineKeyboard(bot, update,
				[][]tb.InlineKeyboardButton{
					[]tb.InlineKeyboardButton{
						tb.NewInlineKeyboardButtonData("Город", "adminCity"),
					},
					[]tb.InlineKeyboardButton{
						tb.NewInlineKeyboardButtonData("Концерты", "adminConcert"),
					},
					[]tb.InlineKeyboardButton{
						tb.NewInlineKeyboardButtonData("Пользователи", "adminUsers"),
					},
					[]tb.InlineKeyboardButton{
						tb.NewInlineKeyboardButtonData("Площадки", "adminVenue"),
					},
				},
				"Админ панель")
		}
	default:
		log.Printf("Unknown command: %s", command)
	}
}

/*
EditInlineKeyboard allow you to edit a sended inline keyboard and text
*/
func EditInlineKeyboard(bot *tb.BotAPI, update *tb.Update, newButtons [][]tb.InlineKeyboardButton, newText string) {
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID
	markup := &tb.InlineKeyboardMarkup{
		InlineKeyboard: newButtons,
	}
	msg := tb.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		newText,
		*markup,
	)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
func handleCallbacks(db *sql.DB, bot *tb.BotAPI, update *tb.Update) {
	clientID := update.CallbackQuery.From.ID
	/*row := db.QueryRow("SELECT login2 FROM public.users WHERE login = $1", clientID)
	if err := row.Scan(&clientID); err != nil {
		log.Panic(err)
	}
	if clientID == 0 {
		_, err := db.Exec("UPDATE public.users SET city $1 WHERE login = $2", clientID)
		if err != nil {
			log.Panic(err)
		}
	}*/
	//messageID := update.CallbackQuery.Message.MessageID
	//clientUserName := update.CallbackQuery.From.UserName
	callbackData := update.CallbackQuery.Data
	switch {
	case strings.HasPrefix(callbackData, "concerts"):

	case strings.HasPrefix(callbackData, "venues"):
		//		parts := strings.SplitN(callbackData, "/", -1)
	//	venue := parts[1]
	//	sendListOf(bot, update, db, parts[0], msg, back, arg, page)

	case strings.HasPrefix(callbackData, "cities"):
		parts := strings.SplitN(callbackData, "/", -1)
		_, err := db.Exec("UPDATE public.users SET citycode = $1 WHERE login = $2", parts[1], clientID)
		if err != nil {
			log.Panic(err)
		}
		row := db.QueryRow("SELECT name FROM public.cities WHERE code = $1", parts[1])
		var city string
		err = row.Scan(&city)
		if err != nil {
			log.Panic(err)
		}
		backMain(bot, update, "Ваш город: "+city)
	case strings.HasPrefix(callbackData, "choose/"):
		parts := strings.SplitN(callbackData, "/", -1)
		var msg, back, arg string
		var page int
		isNextChoose := false
		switch parts[1] {
		case "cities":
			msg = "Выберите город: "
			back = "backMain"
			page, _ = strconv.Atoi(parts[2])
		case "venues":
			msg = "Выберите площадку: "
			back = "chooseConcert"
			page, _ = strconv.Atoi(parts[3])
			arg = " WHERE city = '" + parts[2] + "'"
			isNextChoose = true
		case "concerts":
			msg = "Выберите концерт:"
			back = "chooseConcert"
			if parts[2] != "all" {
				arg = " WHERE venue = '" + parts[2] + "'"
			} else {
				arg = " WHERE city = '" + parts[3] + "'"
			}
		}
		sendListOf(bot, update, db, parts[1], msg, back, arg, page, isNextChoose)
	}
	switch callbackData {
	case "backMain":
		backMain(bot, update, "Добро пожаловать: ")
	case "chooseVenue":
		sendListOf(bot, update, db, "venues", "Выберите площадку: ", "chooseConcert", "", 0, true)
	case "chooseConcertFromAll":
		sendListOf(bot, update, db, "concerts", "Выберите концерт: ", "chooseConcert", "", 0, false)
	case "chooseConcert":
		row := db.QueryRow("SELECT citycode FROM public.users WHERE login = $1", clientID)
		var city string
		err := row.Scan(&city)
		if err != nil {
			log.Panic(err)
		}
		EditInlineKeyboard(bot, update,
			[][]tb.InlineKeyboardButton{
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать из всех", "choose/concerts/all/"+city+"/0"),
				},
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("Выбрать площадку", "choose/venues/"+city+"/0"),
				},
				[]tb.InlineKeyboardButton{
					tb.NewInlineKeyboardButtonData("назад", "backMain"),
				},
			},
			"Выбрать концерт:")
	default:
		log.Printf("%s Unknown command: %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
	}
}

func SendInlineKeyboard(bot *tb.BotAPI, update *tb.Update, newButtons [][]tb.InlineKeyboardButton, newText string) {

	markup := &tb.InlineKeyboardMarkup{
		InlineKeyboard: newButtons,
	}
	msg := tb.NewMessage(update.Message.Chat.ID, newText)
	msg.ReplyMarkup = markup
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

type getEnvError struct{}

func (e getEnvError) Error() string {
	return "Не удалось загрузить .env"
}

func GetEnv(key string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("не удалось загрузить .env", err)
		return "", getEnvError{}
	}
	value := os.Getenv(key)
	return value, nil
}

func isRoot(bot *tb.BotAPI, update *tb.Update, db *sql.DB, login int64, command string) bool {
	isRoot := false
	row := db.QueryRow("SELECT adm FROM public.users WHERE login = $1 ", login)
	err := row.Scan(&isRoot)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(login, " triyng ", command)
		} else {
			log.Panic(err)
		}
	}
	return isRoot
}
func sendListOf(bot *tb.BotAPI, update *tb.Update, db *sql.DB, tableName string, msg string, back string, args string, page int, isNextChoose bool) {
	codes := make([]string, 0)
	names := make([]string, 0)
	keyboard := make([][]tb.InlineKeyboardButton, 0)
	rows, err := db.Query("SELECT code, name FROM public."+tableName+args+" OFFSET $1 LIMIT $2 ", page*4, (page+1)*4)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		var name string
		err := rows.Scan(&code, &name)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			} else {
				log.Panic(err)
			}
		}
		codes = append(codes, code)
		names = append(names, name)
	}
	var prefix string
	if tableName == "venues" {

		if isNextChoose {
			prefix = "choose/"
		}
		for i := 0; i < len(codes); i++ {
			keyboard = append(keyboard, []tb.InlineKeyboardButton{
				tb.NewInlineKeyboardButtonData(names[i], prefix+"concerts"+"/"+codes[i]+"/0"),
			})
		}
	} else {
		for i := 0; i < len(codes); i++ {
			keyboard = append(keyboard, []tb.InlineKeyboardButton{
				tb.NewInlineKeyboardButtonData(names[i], prefix+tableName+"/"+codes[i]),
			})
		}
	}
	keys := make([]tb.InlineKeyboardButton, 0)
	prevPage := strconv.Itoa(page - 1)
	nextPage := strconv.Itoa(page + 1)
	if page != 0 {
		keys = append(keys, tb.NewInlineKeyboardButtonData("<-", "choose/"+tableName+"/"+prevPage))
	}
	keys = append(keys, tb.NewInlineKeyboardButtonData("Назад", back))
	row := db.QueryRow("SELECT code FROM public."+tableName+args+" OFFSET $1 LIMIT $2 ", (page+1)*4, (page+1)*4+1)
	var temp string
	if err := row.Scan(&temp); err == nil {
		keys = append(keys, tb.NewInlineKeyboardButtonData("->", "choose/"+tableName+"/"+nextPage))
	}
	keyboard = append(keyboard, keys)
	EditInlineKeyboard(bot, update, keyboard, msg)
}
func backMain(bot *tb.BotAPI, update *tb.Update, msg string) {
	EditInlineKeyboard(bot, update,
		[][]tb.InlineKeyboardButton{
			[]tb.InlineKeyboardButton{
				tb.NewInlineKeyboardButtonData("Выбрать город", "choose/cities/0"),
			},
			[]tb.InlineKeyboardButton{
				tb.NewInlineKeyboardButtonData("Выбрать концерт", "chooseConcert"),
			},
		},
		msg,
	)
}
