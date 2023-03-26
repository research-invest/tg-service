package main

import (
	"encoding/json"
	"os"
)

var appConfig Config

type Config struct {
	Port     int
	Host     string
	Channels []Channel
	Ssl      Ssl
}

type Channel struct {
	Name    string `json:"name"`
	UrlCode string `json:"url-code"`
	UrlApi  string `json:"url-api"`
	Token   string `json:"token"`
}

type Ssl struct {
	CertificateFile    string `json:"crt-file"`
	CertificateKeyFile string `json:"key-file"`
}

func readConfig() {
	file, err := os.Open("config.json")

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	appConfig = *(new(Config))
	err = decoder.Decode(&appConfig)
	if err != nil {
		panic(err)
	}
}
