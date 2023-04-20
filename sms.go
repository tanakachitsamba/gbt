package main

import (
	"fmt"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

func sendSMS(from, to, message string) {
	client := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody(message)
	params.SetFrom(from)
	params.SetTo(to)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println("Error sending SMS:", err.Error())
	} else {
		if resp.Sid != nil {
			fmt.Println("SMS sent with SID:", *resp.Sid)
		} else {
			fmt.Println("SMS sent with SID:", resp.Sid)
		}
	}
}
