package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//struct for weather date

var yandex Yandex

//token vars

var tgToken string
var yaToken string

var bot *tba.BotAPI

// var to store addtime state of every user separately
var addtimeUsers = make(map[int64]addtimeInfo)

type addtimeInfo struct {
	state      int
	remindTime string
	remindType string
}

//user channel in key
//addtime state[0], time[1], type[2] in value

//state
//0 means not in add time mode
//1 means user entering time
//2 means user choosing type

//keyboard here to reduce space in cycle

var replyKeyboard = tba.NewReplyKeyboard(
	tba.NewKeyboardButtonRow(
		tba.NewKeyboardButton("погода"),
		tba.NewKeyboardButton("прогноз"),
	),
	tba.NewKeyboardButtonRow(
		tba.NewKeyboardButton("погода и прогноз"),
	),
)

//define loggers

var infoLog = log.New(os.Stdout, "INFO # ", log.Ldate|log.Ltime)
var warnLog = log.New(os.Stdout, "WARN # ", log.Ldate|log.Ltime)
var errLog = log.New(os.Stdout, "ERR  # ", log.Ldate|log.Ltime)

//
//
//

func main() {
	infoLog.Println("entered main")
	//get api tokens to begin bot work
	err := getTokens()
	if err != nil {
		//no work if no tokens
		errLog.Fatalln("getTokens returned error:", err)
	}
	infoLog.Println("got api tokens")

	//get coords to begin api work
	err = getCoords()
	if err != nil {
		//no work if no coords
		errLog.Fatalln("getCoords returned error:", err)
	}
	infoLog.Println("got coords")

	//start bot
	bot, err = tba.NewBotAPI(tgToken)
	if err != nil {
		//service unavailable or token incorrect
		errLog.Fatalln("NewBotAPI returned error:", err)
	}
	infoLog.Println("bot started")

	//trying to load schedule data from json
	err = loadSchedule()
	if err != nil {
		if os.IsNotExist(err) {
			//file wil be created and filled during the work
			warnLog.Println("loadSchedule returned error:", err, "# using nil data")
		} else {
			//file exists but dat acant be read
			errLog.Println("loadSchedule returned error:", err)
		}
	} else {
		infoLog.Println("loaded schedule")
	}

	//set time checker to automatically send messages
	//when it is requested time
	go timeChecker()

	//start channel to start getting updates
	u := tba.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	infoLog.Println("starting to get updates")
	//main loop
	for update := range updates {
		logUsername := update.Message.From.UserName + " #"
		infoLog.Println(logUsername, "received message:", update.Message.Text)
		//stores message text
		var text string
		//reply keyboard placeholder
		var replyKb tba.ReplyKeyboardMarkup
		//placeholder for addtimeInfo
		var addtimeInfoEl addtimeInfo

		//if user entered command
		if update.Message.IsCommand() {
			//clean users addtime info
			delete(addtimeUsers, update.Message.Chat.ID)

			//case with command patterns
			switch update.Message.Command() {
			/////////////////////////////////////////////////////////////
			//onstart message
			case "start":
				infoLog.Println(logUsername, "processing as command /start")
				text = "введи <b>/help</b> чтобы получить справочную информацию."

			/////////////////////////////////////////////////////////////
			//should say what bot does
			case "help":
				infoLog.Println(logUsername, "processing as command /help")
				text = "я вывожу информацию о погоде на данный момент <i>(команда <b>/now</b>)</i>" +
					" и её прогнозе <i>(команда <b>/forecast</b>)</i>.\nтакже," +
					" я могу самостоятельно отправлять тебе желаемую информацию ежедневно в указанное тобой время <i>(команда <b>/addtime</b>)</i>.\n" +
					"ты можешь удалить созданное расписание <i>командой <b>/deltime</b></i>."

			/////////////////////////////////////////////////////////////
			//get weather status for now
			case "now":
				infoLog.Println(logUsername, "processing as command /now")
				//update data to send actual information
				infoLog.Println(logUsername, "updating yandex data")
				err := updateYaData()
				if err != nil {
					//something bad happened if cant update data
					errLog.Fatalln("updateYaData returned error:", err)
				}

				//call function for current weather as requested
				text = formatYaWeather()
				infoLog.Println(logUsername, "filled message text with weather data")

			/////////////////////////////////////////////////////////////
			//get weather status for later
			case "forecast":
				infoLog.Println(logUsername, "processing as command /forecast")
				//same as in "now" section
				infoLog.Println(logUsername, "updating yandex data")
				err := updateYaData()
				if err != nil {
					//something bad happened if cant update data
					errLog.Fatalln("updateYaData returned error:", err)
				}

				//call function for current forecast as requested
				text = formatYaForecast()
				infoLog.Println(logUsername, "filled message text with forecast data")

			/////////////////////////////////////////////////////////////
			//enter mode where user sets time for schedule
			case "addtime":
				infoLog.Println(logUsername, "processing as command /addtime")
				addtimeInfoEl.state = 1
				text = "напиши время, в которое хочешь получать информацию о погоде.\n" +
					"новое расписание перезапишет старое.\n" +
					"любая команда отменит запрос\n\nвводи время в виде ЧЧ:ММ."

			/////////////////////////////////////////////////////////////
			//delete users schedule time
			case "deltime":
				infoLog.Println(logUsername, "processing as command /deltime")
				//delete schedule by users channel
				err = delFromSchedule(update.Message.Chat.ID)
				if err != nil {
					//specific reply if users has not schedule
					if errors.Is(err, errUserNotFound) {
						text = "расписание для удаления не найдено"
						infoLog.Println(logUsername, "time to delete not found")
					} else {
						//else something unexpected :(
						errLog.Fatalln("delFromSchedule returned error:", err)
					}
				} else {
					//else success
					text = "расписание удалено"
					infoLog.Println(logUsername, "time deleted")
				}

			/////////////////////////////////////////////////////////////
			//if nothing matches
			default:
				infoLog.Println(logUsername, "command not found:", update.Message.Text)
				text = "такой команды не существует"
			}

		} else if update.Message.Text != "" {
			//if user sent simple text

			//load existing data if user already has
			addtimeInfoEl = addtimeUsers[update.Message.Chat.ID]

			if addtimeInfoEl.state == 1 {
				infoLog.Println(logUsername, "processing as addtime state 1")
				//if user sets his schedule now

				//pick time suggested by user
				schedReq := update.Message.Text

				//check if time in correct format
				if isTimeCorrect(schedReq) {
					infoLog.Println(logUsername, "received valid time:", schedReq)
					//if yes then fill in users info storage
					addtimeInfoEl.remindTime = schedReq

					//advance to next stage
					text = "выбери вид информации, которую ты хочешь получать по расписанию"
					replyKb = replyKeyboard

					addtimeInfoEl.state = 2
					infoLog.Println(logUsername, "advanced to addtime state 2")
				} else {
					infoLog.Println(logUsername, "received invalid time:", schedReq, "# waiting for valid")
					//else send signal and give user another try
					text = "неправильный вид"
				}

			} else if addtimeInfoEl.state == 2 {
				infoLog.Println(logUsername, "processing as addtime state 2")
				//if user chooses his remind type now

				//pick remind type suggested by user
				schedType := strings.ToLower(update.Message.Text)

				//check if type has one of possible values
				if schedType == "погода" || schedType == "прогноз" || schedType == "погода и прогноз" {
					infoLog.Println(logUsername, "received valid type:", schedType)
					switch schedType {
					case "погода":
						schedType = "w"
					case "прогноз":
						schedType = "f"
					case "погода и прогноз":
						schedType = "a"
					}
					//add remind type to struct
					addtimeInfoEl.remindType = schedType
					text = "расписание добавлено"

					infoLog.Println(logUsername,
						"adding user to schedule with chat id:", update.Message.Chat.ID,
						"remind time:", addtimeInfoEl.remindTime,
						"remind type:", addtimeInfoEl.remindType,
					)
					//add correct time to bot schedule list
					err = addToSchedule(update.Message.Chat.ID, addtimeInfoEl.remindTime, addtimeInfoEl.remindType)
					if err != nil {
						//if error maybe data broken
						//better reenter
						text = "добавить расписание не удалось"
						infoLog.Println(logUsername, "failed to add user to schedule")
					} else {
						infoLog.Println(logUsername, "user added to schedule")
					}
					//set state to 0, user leaves addtime mode
					addtimeInfoEl.state = 0
					//delete user from memory
					delete(addtimeUsers, update.Message.Chat.ID)
					infoLog.Println(logUsername, "leaving addtime state")

				} else {
					infoLog.Println(logUsername, "received invalid type:", schedType, "# waiting for valid")
					//send signal and give user another try
					text = "неправильный тип"
				}

			} else {
				//else dont answer
				continue
			}
		} else {
			//else dont answer
			continue
		}

		//update addtimeUsers map with worked on struct
		//if user is in addtime mode
		if addtimeInfoEl.state != 0 {
			infoLog.Println(logUsername, "updating addtime info:", addtimeInfoEl)
			addtimeUsers[update.Message.Chat.ID] = addtimeInfoEl
		}

		//send message and report unsuccessful try
		msg, err := sendMessage(text, update.Message.Chat.ID, replyKb)
		if err != nil {
			//log excuse
			warnLog.Printf(logUsername, "answer message not sent: %v\n%#v\n", err, msg)
		} else {
			infoLog.Println(logUsername, "answer message sent")
		}

	}
}

// sends message with given text to given channel.
// returns full message config if error
func sendMessage(text string, channel int64, replyKb tba.ReplyKeyboardMarkup) (tba.Message, error) {
	//new message config
	msgConf := tba.NewMessage(channel, text)
	msgConf.ParseMode = "HTML"
	msgConf.DisableWebPagePreview = true

	//if received not empty keuboard
	if len(replyKb.Keyboard) != 0 {
		//show it
		replyKb.OneTimeKeyboard = true
		msgConf.ReplyMarkup = replyKb
	} else {
		//else close it
		msgConf.ReplyMarkup = tba.NewRemoveKeyboard(true)
	}

	//send created message
	msg, err := bot.Send(msgConf)
	if err != nil {
		//not bad
		return msg, err
	}
	//actually good
	return msg, nil
}

// gets api tokens from file
func getTokens() error {
	//at first open file with tokens inside
	file, err := os.Open("workdata/tokens.txt")
	if err != nil {
		//cant even work without tokens
		return err
	}
	//scanner has every file line as separated element
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		switch i {
		case 2:
			//always tg token on the second line
			tgToken = scanner.Text()
		case 4:
			//always ya token on the fourth line
			yaToken = scanner.Text()
		}
	}
	//close file after work
	err = file.Close()
	if err != nil {
		//single not closed file is not a problem
		warnLog.Println("failed to close file:", err)
	}
	//function ended successfully
	return nil
}
