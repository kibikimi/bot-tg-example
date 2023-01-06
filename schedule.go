package main

//stores functions related to schedule reminder

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// works parallel to main.
// every minute searches for users
// who need message in specified time
func timeChecker() {
	infoLog.Println("entered timechecker")
	var emptyKb tba.ReplyKeyboardMarkup
	var now time.Time
	//ends only with program end
	for {
		//define time vars for further work
		currentTime := now.Truncate(time.Minute).Add(time.Minute)
		currentHour := currentTime.Hour()
		currentMinute := currentTime.Minute()

		//for every time in timetable
		for timeStr, users := range timeTable {
			//convert key part to hour int
			remindHour, err := strconv.Atoi(string([]rune(timeStr)[:2]))
			if err != nil {
				//cant check time and send messages if cant read timetable
				errLog.Fatalln("error in timeChecker:", err)
			}
			//convert key part to minute int
			remindMinute, err := strconv.Atoi(string([]rune(timeStr)[3:]))
			if err != nil {
				//cant check time and send messages if cant read timetable
				errLog.Fatalln("error in timeChecker:", err)
			}

			//if users want remind now
			if currentHour == remindHour && currentMinute == remindMinute {
				infoLog.Printf("remind time %d:%d\n", currentHour, currentMinute)
				infoLog.Println("updating yandex data for reminder")
				//update data to send actual information to users
				err := updateYaData()
				if err != nil {
					//something bad happened if cant update data
					errLog.Fatalln("updateYaData returned error:", err)
				}

				//send message to every user who wants remind now
				for _, userData := range users {
					infoLog.Println("reminding to:", userData)
					var text string
					//choose message type to send
					switch userData.Type {
					case "w":
						text = formatYaWeather()
					case "f":
						text = formatYaForecast()
					case "a":
						text = formatYaWeather() + "\n\n" + formatYaForecast()
					}
					//send message and report unsuccessful try
					msg, err := sendMessage(text, userData.Channel, emptyKb)
					if err != nil {
						//log excuse
						warnLog.Printf("remind message not sent: %v\n%#v\n", err, msg)
					} else {
						infoLog.Println("remind message sent")
					}
				}
			} else {
				infoLog.Printf("not remind time %d:%d\n", currentHour, currentMinute)
			}
		}
		now = time.Now()
		//count time to wait to next minute
		time.Sleep(now.Truncate(time.Minute).Add(time.Minute).Sub(now))
	}
}

// checks given time in format HH:MM
func isTimeCorrect(timeStr string) bool {
	//length must be 5 and 3rd symbol is :
	if len(timeStr) != 5 || string([]rune(timeStr)[2]) != ":" {
		return false
	}

	//if given 1,2,4 or 5 symbols are not digits
	//atoi will return error

	//first number musst be < 24
	num, err := strconv.Atoi(string([]rune(timeStr)[:2]))
	if err != nil || num > 23 {
		return false
	}
	//second number must be < 60
	num, err = strconv.Atoi(string([]rune(timeStr)[3:]))
	if err != nil || num > 59 {
		return false
	}
	//if no errors then time is correct
	return true
}

// adds given time to users schedule
// and replaces old schedule if existed before
func addToSchedule(channel int64, timeStr string, remType string) error {
	//fill userData element with given information
	var userDataEl UserData
	userDataEl.Channel = channel
	userDataEl.Type = remType
	//if user had schedule before
	//he will be deleted before new adding
	err := delFromSchedule(channel)
	if err != nil && !errors.Is(err, errUserNotFound) {
		return err
	}
	//create user schedule
	timeTable[timeStr] = append(timeTable[timeStr], userDataEl)

	err = saveSchedule()
	if err != nil {
		return err
	}
	//finished without errors
	return nil
}

// deletes schedule of given user
func delFromSchedule(channel int64) error {
	var found bool
	//for every time in timetable
	for timeStr, userdataSlice := range timeTable {
		//for every slice with userdata in time
		for i, user := range userdataSlice {
			//if found requested user
			if user.Channel == channel {
				//if user single in time
				if len(userdataSlice) == 1 {
					//then delete whole time
					delete(timeTable, timeStr)
				} else {
					//else there is other users
					//delete only certain user
					if i == 0 {
						//if element first
						timeTable[timeStr] = userdataSlice[1:]
					} else if i+1 == len(userdataSlice) {
						//if element last
						timeTable[timeStr] = userdataSlice[:i]
					} else {
						//else in the middle
						timeTable[timeStr] = append(userdataSlice[:i], userdataSlice[i+1:]...)
					}
				}
				//leave inner cycle
				found = true
				break
			}
		}
	}
	if !found {
		//if user not found
		return errUserNotFound
	}
	//save updated timetable
	err := saveSchedule()
	if err != nil {
		return err
	}
	//good delete and save
	return nil
}

// saves current timtable state to json file
func saveSchedule() error {
	//marshal timetable to json
	file, err := json.Marshal(timeTable)
	if err != nil {
		return err
	}
	//save timetable to file
	err = os.WriteFile("workdata/schedule.json", file, 0644)
	if err != nil {
		return err
	}
	//well saved
	infoLog.Println("saved schedule")
	return nil
}

// loads schedule data from json to timetable
func loadSchedule() error {
	//trying to read json file
	file, err := os.ReadFile("workdata/schedule.json")
	if err != nil {
		//if err then nothing to load
		return err
	}
	//unmarshalling json data to timetable map
	err = json.Unmarshal(file, &timeTable)
	if err != nil {
		//very bad if file exists
		//but data cant be loaded
		return err
	}
	//loaded successfully!
	return nil
}

//////////////////////////////////////////////////////////////////////

var errUserNotFound = errors.New("user to delete not found")

// stores all schedule information
var timeTable = make(map[string][]UserData)

type UserData struct {
	Channel int64
	Type    string
}
