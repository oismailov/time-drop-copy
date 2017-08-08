package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"timedrop/helpers"

	"github.com/timehop/apns"
)

var apnsCert string
var apnsKey string

// PushToken for APN and GCM
type PushToken struct {
	BaseModel

	Token     string `json:"token" valid:"required" sql:"unique;"`
	Platform  string `json:"platform" valid:"required"`
	UserRefer uint   `json:"-"`
}

// PushNotification as helper class
type PushNotification struct {
}

// ReadCerts returns new APNS clients
func (pushNotification *PushNotification) ReadCerts() error {
	certPrefix := "development_"
	if os.Getenv("DEV") == "" {
		certPrefix = "production_"
	}

	if apnsCert == "" {
		apnsPEM, apnsPEMErr := ioutil.ReadFile("assets/certs/" + certPrefix + "com.faktorzwei.puzzle.Time-Drop.pem")
		if apnsPEMErr != nil {
			return apnsPEMErr
		}
		apnsCert = string(apnsPEM)
	}
	if apnsKey == "" {
		apnsPkey, apnsKeyErr := ioutil.ReadFile("assets/certs/" + certPrefix + "com.faktorzwei.puzzle.Time-Drop.pkey")
		if apnsKeyErr != nil {
			return apnsKeyErr
		}
		apnsKey = string(apnsPkey)
	}
	return nil
}

// GetNewAPNSClient returns new APNS clients
func (pushNotification *PushNotification) GetNewAPNSClient() (apns.Client, error) {
	pushNotification.ReadCerts()

	apnsGateway := apns.SandboxGateway
	if os.Getenv("DEV") == "" {
		apnsGateway = apns.ProductionGateway
	}

	client, err := apns.NewClient(apnsGateway, apnsCert, apnsKey)
	if err != nil {
		return apns.Client{}, err
	}
	return client, nil
}

//SendFriendRequestPush sends friend request push
func (pushNotification PushNotification) SendFriendRequestPush(receiver User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	p.APS.Alert.Body = helpers.TranslateStr("push_friend_request_received", receiver.Language)
	p.APS.ContentAvailable = 1

	for _, pushToken := range receiver.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}

//SendFriendRequestAcceptedPush sends friend request push
func (pushNotification PushNotification) SendFriendRequestAcceptedPush(requester User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	p.APS.Alert.Body = helpers.TranslateStr("push_friend_request_accepted", requester.Language)
	p.APS.ContentAvailable = 1

	for _, pushToken := range requester.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}

//SendGameRequestPush sends friend request push
func (pushNotification PushNotification) SendGameRequestPush(receiver User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	p.APS.Alert.Body = helpers.TranslateStr("push_game_challenge_received", receiver.Language)
	p.APS.ContentAvailable = 1

	for _, pushToken := range receiver.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}

//SendGameWonPush sends friend request push
func (pushNotification PushNotification) SendGameWonPush(receiver User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	p.APS.Alert.Body = helpers.TranslateStr("push_game_won", receiver.Language)
	p.APS.ContentAvailable = 1

	for _, pushToken := range receiver.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}

//SendGameLostPush sends friend request push
func (pushNotification PushNotification) SendGameLostPush(receiver User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	p.APS.Alert.Body = helpers.TranslateStr("push_game_lost", receiver.Language)
	p.APS.ContentAvailable = 1

	for _, pushToken := range receiver.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}

//

//SendLifeRequestPush
func (pushNotification PushNotification) SendLifeRequestPush(requester User, userName string) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	message := fmt.Sprintf(helpers.TranslateStr("got_life_request", requester.Language), userName)
	p.APS.Alert.Body = message
	p.APS.ContentAvailable = 1
	fmt.Printf("NewPayload:=====%+v\n\n", p)
	for _, pushToken := range requester.GetAPNSTokens() {
		fmt.Printf("pushToken:===%+v\n\n", pushToken)
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}
	f, err := apns.NewFeedback(apns.ProductionGateway, apnsCert, apnsKey)
	if err != nil {
		log.Fatal("Could not create feedback", err.Error())
	}
	for ft := range f.Receive() {
		fmt.Println("Feedback for token:", ft.DeviceToken)
	}

	return nil
}

//GiveLifeRequestPush
func (pushNotification PushNotification) GiveLifeRequestPush(requester User) (err error) {
	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	message := helpers.TranslateStr("got_life", requester.Language)
	p.APS.Alert.Body = message
	p.APS.ContentAvailable = 1

	for _, pushToken := range requester.GetAPNSTokens() {
		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = pushToken.Token
		m.Priority = apns.PriorityImmediate

		err := apnsClient.Send(m)
		fmt.Println(err)
	}

	return nil
}
