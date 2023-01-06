package main

//stores all the format functions

import (
	"fmt"
	"strconv"
	"time"
)

// format time int parts as single string
func FormatTime(unixDate time.Time) string {
	hour := strconv.Itoa(unixDate.Hour())
	minute := strconv.Itoa(unixDate.Minute())
	//minute always should be 2 symbols in length
	if len(minute) == 1 {
		minute = "0" + minute
	}

	return hour + ":" + minute
}

//////////////////////////////////////////////////////////////////////

// change date ints to their translated string equivalents
func FormatDate(unixDate time.Time) string {
	day := unixDate.Day()
	monthInt := int(unixDate.Month())
	weekInt := int(unixDate.Weekday())

	var month string
	switch monthInt {
	case 1:
		month = "января"
	case 2:
		month = "февраля"
	case 3:
		month = "марта"
	case 4:
		month = "апреля"
	case 5:
		month = "мая"
	case 6:
		month = "июня"
	case 7:
		month = "июля"
	case 8:
		month = "августа"
	case 9:
		month = "сентября"
	case 10:
		month = "октября"
	case 11:
		month = "ноября"
	case 12:
		month = "декабря"
	default:
		warnLog.Println("unexpected monthInt value:", monthInt)
		month = "_"
	}

	var weekDay string
	switch weekInt {
	case 1:
		weekDay = "понедельник"
	case 2:
		weekDay = "вторник"
	case 3:
		weekDay = "среда"
	case 4:
		weekDay = "четверг"
	case 5:
		weekDay = "пятница"
	case 6:
		weekDay = "суббота"
	case 0:
		weekDay = "воскресенье"
	default:
		warnLog.Println("unexpected weekInt value:", weekInt)
		weekDay = "_"
	}

	return strconv.Itoa(day) + " " + month + ", " + weekDay
}

//////////////////////////////////////////////////////////////////////

// translation of condition
func FormatCondition(condition string) string {
	switch condition {
	case "clear":
		condition = "ясно"
	case "partly-cloudy":
		condition = "малооблачно"
	case "cloudy":
		condition = "облачно с прояснениями"
	case "overcast":
		condition = "пасмурно"
	case "drizzle":
		condition = "морось"
	case "light-rain":
		condition = "небольшой дождь"
	case "rain":
		condition = "дождь"
	case "moderate-rain":
		condition = "умеренный дождь"
	case "heavy-rain":
		condition = "сильный дождь"
	case "continuous-heavy-rain":
		condition = "длительный сильный дождь"
	case "showers":
		condition = "ливень"
	case "wet-snow":
		condition = "дождь со снегом"
	case "light-snow":
		condition = "небольшой снег"
	case "snow":
		condition = "снег"
	case "snow-showers":
		condition = "снегопад"
	case "hail":
		condition = "град"
	case "thunderstorm":
		condition = "гроза"
	case "thunderstorm-with-rain":
		condition = "дождь с грозой"
	case "thunderstorm-with-hail":
		condition = "гроза с градом"
	default:
		warnLog.Println("unexpected condition value:", condition)
		condition = "_"
	}

	return condition
}

//////////////////////////////////////////////////////////////////////

// decoding and translation of windDir
func FormatWindDir(windDir string) string {
	switch windDir {
	case "nw":
		windDir = "северо-западный ветер"
	case "n":
		windDir = "северный ветер"
	case "ne":
		windDir = "северо-восточный ветер"
	case "e":
		windDir = "восточный ветер"
	case "se":
		windDir = "юго-восточный ветер"
	case "s":
		windDir = "южный ветер"
	case "sw":
		windDir = "юго-западный ветер"
	case "w":
		windDir = "западный ветер"
	case "c":
		windDir = "штиль"
	default:
		warnLog.Println("unexpected windDir value:", windDir)
		windDir = "_"
	}

	return windDir
}

//////////////////////////////////////////////////////////////////////

// decoding and translation of dayTime
func FormatDayTime(dayTime string) string {
	switch dayTime {
	case "d":
		dayTime = "светло"
	case "n":
		dayTime = "темно"
	default:
		warnLog.Println("unexpected dayTime value:", dayTime)
		dayTime = "_"
	}

	return dayTime
}

//////////////////////////////////////////////////////////////////////

// translation and declension of dayPart
func FormatDayPart(dayPart string) string {
	switch dayPart {
	case "night":
		dayPart = "ночью"
	case "morning":
		dayPart = "утром"
	case "day":
		dayPart = "днём"
	case "evening":
		dayPart = "вечером"
	default:
		warnLog.Println("unexpected dayPart value:", dayPart)
		dayPart = "_"
	}

	return dayPart
}

//////////////////////////////////////////////////////////////////////

// gets current weather info from data struct and returns it formatted
func formatYaWeather() string {
	//pick timestamp of query
	unixDate := time.Unix(yandex.UnixTime, 0)
	//get formatted date
	queryDate := FormatDate(unixDate)
	//get formatted time
	queryTime := FormatTime(unixDate)

	//get formatted condition
	condition := FormatCondition(yandex.Fact.Condition)
	//get formatted daytime
	dayTime := FormatDayTime(yandex.Fact.Daytime)
	//get formatted wind direction
	windDir := FormatWindDir(yandex.Fact.WindDir)

	//unite weather parameters in formatted string
	text := fmt.Sprintf(
		"<i><a href=\"%s\">погода</a> на %s\n"+ //////Url and queryTime
			"%s\n\n</i>"+ ////////////////////////////queryDate
			"на улице <b>%s</b> и <b>%s</b>\n"+ //////condition and dayTime
			"<b>%.f°</b>, ощущается <b>%.f</b>°\n"+ //Temp and TempFeels
			"%s, <b>%.1f</b> м/с\n"+ /////////////////windDir and WindSpeed
			"порывы до <b>%.1f</b> м/с\n"+ ///////////GustSpeed
			"давление <b>%.f</b> мм.рт.ст.\n"+ ///////PressureMm
			"влажность <b>%.f%%</b>\n", //////////////Humidity

		yandex.Info.Url,
		queryTime,
		queryDate,
		condition, dayTime,
		yandex.Fact.Temp, yandex.Fact.TempFeels,
		windDir, yandex.Fact.WindSpeed,
		yandex.Fact.GustSpeed,
		yandex.Fact.PressureMm,
		yandex.Fact.Humidity,
	)
	//return string with current weather information
	return text
}

//////////////////////////////////////////////////////////////////////

// gets current forecast info from data struct and returns it formatted
func formatYaForecast() string {
	//pick timestamp of query
	unixDate := time.Unix(yandex.Forecast.UnixTime, 0)
	//get formatted date
	queryDate := FormatDate(unixDate)

	//get sunrise and sunset time
	sunrise := yandex.Forecast.Sunrise
	sunset := yandex.Forecast.Sunset

	//message text placeholder with onetime displayed data
	text := fmt.Sprintf(
		"<i><a href=\"%s\">прогноз</a> на %s\n"+ //Url and queryDate
			"восход в %s - закат в %s\n</i>", /////sunrise and sunset

		yandex.Info.Url, queryDate,
		sunrise, sunset,
	)

	//forecast parts
	part1 := yandex.Forecast.Part1
	part2 := yandex.Forecast.Part2

	//make array from parts and read them in order
	for _, part := range [2]Part{part1, part2} {
		//getting and formatting data in single line
		dayPart := FormatDayPart(part.DayPart)
		condition := FormatCondition(part.Condition)
		dayTime := FormatDayTime(part.Daytime)
		windDir := FormatWindDir(part.WindDir)

		//add to message
		text += fmt.Sprintf(
			"\n<u><b>%s</b> на улице</u>\n"+ /////////////dayPart
				"<b>%s</b> и <b>%s</b>\n"+ ///////////////condition and dayTime
				"<b>%.f°</b>, ощущается <b>%.f</b>°\n"+ //Temp and TempFeels
				"%s, <b>%.1f</b> м/с\n"+ /////////////////windDir and WindSpeed
				"порывы до <b>%.1f</b> м/с\n"+ ///////////GustSpeed
				"давление <b>%.f</b> мм.рт.ст.\n"+ ///////PressureMm
				"влажность <b>%.f%%</b>\n", //////////////Humidity

			dayPart,
			condition, dayTime,
			part.Temp, part.TempFeels,
			windDir, part.WindSpeed,
			part.GustSpeed,
			part.PressureMm,
			part.Humidity,
		)
	}
	//return string with forecast information
	return text
}
