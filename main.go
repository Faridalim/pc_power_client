package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tkanos/gonfig"
)

// default more than 15 so it become false/unused at app start
var timeTracked int = 20

type Configuration struct {
	Id     string
	Remote string
}

type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Id          int       `json:"id"`
		Last_active time.Time `json:"last_active"`
		Perintah    string    `json:"perintah"`
	} `json:"data"`
}

func main() {

	// loop last active
	go updateLastActive()

	//loop time tracker
	go timeTracker()
	//run local server to receive if app is used
	e := echo.New()
	e.GET("/", reportState)
	e.GET("/cek", getTracked)
	e.Logger.Fatal(e.Start("127.0.0.1:2000"))

}

func readConfig() Configuration {
	conf := Configuration{}
	gonfig.GetConf("config.json", &conf)
	return conf
}

func updateLastActive() {
	for {
		conf := readConfig()
		targetUrl := "http://" + conf.Remote + ":3001/pc/last_active"
		jsonResult := Result{}

		// check apakah pc in use/tidak. Jika time tracked > 10 maka false. if time tracked <= 10 maka true

		var pcInUse bool = timeTracked <= 15
		var strInUse string
		if pcInUse {
			strInUse = "true"
		} else {
			strInUse = "false"
		}

		data := url.Values{}
		data.Set("id", conf.Id)
		data.Set("in_use", strInUse)

		client := &http.Client{}
		req, err := http.NewRequest("POST", targetUrl, strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Println("Request string error")
			continue
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("COntent-Type", strconv.Itoa(len(data.Encode())))

		res, err := client.Do(req)
		if err != nil {
			fmt.Println("Request to remote failed")
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Read Result failed")
			continue
		}

		res.Body.Close()

		// fmt.Println(string(body))

		err = json.Unmarshal(body, &jsonResult)
		if err != nil {
			fmt.Println(err)
		}

		if jsonResult.Data.Perintah == "shutdown" {
			err := exec.Command("cmd", "/C", "shutdown", "/s").Run()
			if err != nil {
				fmt.Println("Shutdown failed")
				continue
			}
		} else if jsonResult.Data.Perintah == "restart" {
			err := exec.Command("cmd", "/C", "shutdown", "/r").Run()
			if err != nil {
				fmt.Println("Restart Failed")
				continue
			}
		}

		fmt.Println("last_active : ", jsonResult.Data.Last_active)
		time.Sleep(5 * time.Second)
	}
}

func reportState(c echo.Context) error {
	timeTracked = 0
	fmt.Println("time tracked : ", timeTracked)
	return c.JSON(200, map[string]string{
		"message": "sukses",
	})
}

func getTracked(c echo.Context) error {
	return c.JSON(200, map[string]int{
		"time tracker": timeTracked,
	})
}

func timeTracker() {
	for {
		time.Sleep(1 * time.Second)
		timeTracked = timeTracked + 1
		fmt.Println("time tracker : ", timeTracked)
	}
}
