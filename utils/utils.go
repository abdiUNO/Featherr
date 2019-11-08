package utils

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"log"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   error
}

type ErrorResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

const (
	UPDATE_FRIENDS       = "UPDATE_FRIENDS"       // action cannot be performed
	UPDATE_CLIQUES       = "UPDATE_CLIQUES"       // action cannot be performed
	UPDATE_CONVERSATIONS = "UPDATE_CONVERSATIONS" // action cannot be performed
)

type MessageData struct {
	MsgType string `json:"type"`
	Topic   string `json:"topic"`
	UserId  string `json:"userId"`
}

func SendToToic(data *MessageData) {
	ctx := context.Background()

	app, appErr := firebase.NewApp(ctx, nil)
	if appErr != nil {
		log.Fatalln(appErr)
	}

	fcm, fcmErr := app.Messaging(ctx)
	if fcmErr != nil {
		log.Fatalln(fcmErr)
	}

	fmt.Println("Creating Message")

	m := messaging.Message{
		Data: map[string]string{
			"updated": "true",
			"type":    data.MsgType,
			"topic":   data.Topic,
			"userId":  data.UserId,
		},
		Topic: data.Topic,
	}

	fmt.Println("Sending Message")

	resp, err := fcm.Send(ctx, &m)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Successfully sent message:", resp)

	return
}
