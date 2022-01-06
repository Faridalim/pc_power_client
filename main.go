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

	"github.com/tkanos/gonfig"
)

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
		Perintah    string    `json: "perintah"`
	} `json:"data"`
}

func main() {
	for {
		conf := readConfig()
		targetUrl := "http://" + conf.Remote + ":3001/pc/last_active"
		jsonResult := Result{}

		data := url.Values{}
		data.Set("id", conf.Id)

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

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Read Result failed\n")
			continue
		}

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

func readConfig() Configuration {
	conf := Configuration{}
	gonfig.GetConf("config.json", &conf)
	return conf
}
