package main

//stores structs with yandex api information
//and related to them functions

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

// updates yandex data with new info if previous data is old
func updateYaData() error {
	//if bot just booted and struct is empty
	//trying to get info from oldloaded jsons
	if yandex.UnixTime == 0 {
		//check if json file exists
		_, err := os.Stat("workdata/yandex.json")
		if err == nil {
			//if file exists and is valid
			err := updateYaStruct()
			if err != nil {
				//all errors from updateYaStruct are bad (cant work)
				return err
			}
		}
	}

	//get current time and time of last update
	lastUpdate := time.Unix(yandex.UnixTime, 0)
	now := time.Now()
	//if last update is old
	//or not in current halfhour part (16:00-16:30;16:30-17:00)
	//then update with new data
	if now.Sub(lastUpdate).Hours() > 12 ||
		lastUpdate.Hour() != now.Hour() ||
		(lastUpdate.Minute() < 30 && now.Minute() >= 30) ||
		(lastUpdate.Minute() >= 30 && now.Minute() < 30) {
		//get new data
		err := getNewYaData()
		if err != nil {
			//if cant get new data
			//check if can use old
			_, err := os.Stat("workdata/yandex.json")
			if err == nil {
				//if file exists and is valid
				warnLog.Println("error in updateYaData:", err, "# using older data")
				//still can work
				return nil
			} else {
				//else cant work with old data
				return err
			}
		}
		//and fill struct with new data
		err = updateYaStruct()
		if err != nil {
			return err
		}
	}
	//function finished without errors
	return nil
}

// gets new json file from server
func getNewYaData() error {
	//make http client
	client := &http.Client{}

	//create request on server
	request, err := http.NewRequest("GET", yaUrl, nil)
	if err != nil {
		//very bad error means url or coords wrong
		//if file already exists use old data
		return err
	}

	//set header with api key
	request.Header.Set("X-Yandex-API-Key", yaToken)
	//make request and accept json
	result, err := client.Do(request)
	if err != nil {
		//wrong api? anyways cant get new data
		//use older data if file exists already
		return err
	}

	//create file to store new data
	file, err := os.Create("workdata/yandex.json")
	if err != nil {
		//why failed to create? maybe previous version is broken now?
		//even if there is no old version file does not exists
		//better stop work
		//OR i think it checks in updateYaData
		//if file is normal
		//anyways it will be noticable
		return err
	}

	//fill file with accepted json
	_, err = io.Copy(file, result.Body)
	if err != nil {
		//same as previous error
		return err
	}
	//did it!!!!!!!!
	infoLog.Printf("got new yandex data")
	return nil
}

// updates struct weather fields with json information
func updateYaStruct() error {
	//read just created or existed before file for json in byte type
	body, err := os.ReadFile("workdata/yandex.json")
	if err != nil {
		//if cant read then cant get data too
		return err
	}

	//load new json to struct (only weather!)
	err = json.Unmarshal(body, &yandex)
	if err != nil {
		//broken data if not marshalling
		//fatallll
		return err
	}

	//load each (forecast!) part separately
	part1 := yandex.Forecast.Parts[0]
	part2 := yandex.Forecast.Parts[1]
	//prepare map to hold maps
	partsMap := make(map[string]map[string]interface{})
	//assign values with keys as struct fields
	partsMap["Part1"] = part1
	partsMap["Part2"] = part2

	//marshal map to json
	partsJson, err := json.Marshal(partsMap)
	if err != nil {
		//need this map to json so bad cant even work further
		return err
	}
	//and unmarshal to struct for automatic field filling
	err = json.Unmarshal(partsJson, &yandex.Forecast)
	if err != nil {
		//howwwww literally unexpected error
		return err
	}
	//function ended successfully
	infoLog.Printf("yandex struct updated")
	return nil
}

// gets coordinates from file
func getCoords() error {
	//at first open file with coords inside
	file, err := os.Open("workdata/coords.txt")
	if err != nil {
		//cant send correct data without coords
		return err
	}
	//scanner has every file line as separated element
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		switch i {
		case 2:
			//always lat on the second line
			yaUrl += "?lat=" + scanner.Text()
		case 4:
			//always lon on the fourth line
			yaUrl += "&lon=" + scanner.Text()
		}
	}
	//close file after work
	err = file.Close()
	if err != nil {
		//second not closed file is not a problem
		warnLog.Println("failed to close file:", err)
	}
	//function ended successfully
	return nil
}

//////////////////////////////////////////////////////////////////////

var yaUrl string = "https://api.weather.yandex.ru/v2/informers"

type Yandex struct {
	UnixTime int64 `json:"now"`
	Info     Info
	Fact     Fact
	Forecast Forecast
	//Date     string `json:"now_dt"`
}

type Info struct {
	Url string `json:"url"`
	//Lat float64
	//Lon float64
}

type Fact struct {
	Temp       float64
	TempFeels  float64 `json:"feels_like"`
	Icon       string
	Condition  string
	WindSpeed  float64 `json:"wind_speed"`
	WindDir    string  `json:"wind_dir"`
	GustSpeed  float64 `json:"wind_gust"`
	PressureMm float64 `json:"pressure_mm"`
	Humidity   float64
	Daytime    string
	//InfoTime   int64 `json:"obs_time"`
	//TempWater  float64 `json:"temp_water"`
	//PressurePa float64 `json:"pressure_pa"`
	//Polar      bool
}

type Forecast struct {
	UnixTime int64 `json:"date_ts"`
	Sunrise  string
	Sunset   string
	Parts    [2]map[string]interface{}
	Part1    Part
	Part2    Part
	//Date string
	//Week int
	//MoonPhase     int    `json:"moon_code"`
	//MoonPhaseText string `json:"moon_text"`
}

type Part struct {
	DayPart    string  `json:"part_name"`
	Temp       float64 `json:"temp_avg"`
	WindSpeed  float64 `json:"wind_speed"`
	WindDir    string  `json:"wind_dir"`
	GustSpeed  float64 `json:"wind_gust"`
	PressureMm float64 `json:"pressure_mm"`
	Humidity   float64
	TempFeels  float64 `json:"feels_like"`
	Condition  string
	Daytime    string
	Icon       string
	//TempMin float64 `json:"temp_min"`
	//TempMax    float64 `json:"temp_max"`
	//PrecMm     float64 `json:"prec_mm"`
	//PrecProb   float64 `json:"prec_prob"`
	//PrecPeriod float64 `json:"prec_period"`
	//PressurePa float64 `json:"pressure_pa"`
	//Polar     bool
}
