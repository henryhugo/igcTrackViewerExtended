package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func clock_trigger(value_to_check int) {

	for {

		if timestamp != value_to_check {
			for _, wh := range whDB {
				text := "{\"text\": \"New track added\"}"
				payload := strings.NewReader(text)
				client := &http.Client{Timeout: (time.Second * 30)}
				req, err := http.NewRequest("POST", wh.WebhookURL, payload)
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					fmt.Print(err.Error())
				}
				fmt.Println(resp.Status)

			}

			time.Sleep(10 * time.Minute)
		}

	}
}
