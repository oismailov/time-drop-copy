package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"timedrop/helpers"

	"github.com/NaySoftware/go-fcm"
	"github.com/timehop/apns"
)

var apnsCert string
var apnsKey string

const (
	firebaseApiKey = "AAAA7_1fgrU:APA91bHJFj-p6ftqlTzedOIi48cYyjTqgbi6AueBxs6lr_s_v8rRzG5gS66fdTCrojsh_MA9Lsi_INcI02tf3cCyNcdpqnYKHrdIaVPcJSaRVYcq72-Pf7ujFB706ODoUSrgnfqcganQ"
)

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

type PushNotificationFCM struct {
	Message string `json:"message"`
	Title   string `json:"title"`
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
	for _, pushToken := range receiver.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("push_friend_request_received", receiver.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", receiver.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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

	for _, pushToken := range requester.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("push_friend_request_accepted", requester.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", requester.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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

	for _, pushToken := range receiver.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("push_game_challenge_received", receiver.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", receiver.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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

	for _, pushToken := range receiver.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("push_game_won", receiver.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", receiver.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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

	for _, pushToken := range receiver.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("push_game_lost", receiver.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", receiver.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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
	for _, pushToken := range requester.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = fmt.Sprintf(helpers.TranslateStr("got_life_request", requester.Language), userName)
		data.Title = helpers.TranslateStr("fcm_push_title", requester.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

	apnsClient, err := pushNotification.GetNewAPNSClient()
	if err != nil {
		return err
	}

	// Create payload
	p := apns.NewPayload()
	message := fmt.Sprintf(helpers.TranslateStr("got_life_request", requester.Language), userName)
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

	for _, pushToken := range requester.GetFireBaseTokens() {
		var data PushNotificationFCM
		data.Message = helpers.TranslateStr("got_life", requester.Language)
		data.Title = helpers.TranslateStr("fcm_push_title", requester.Language)

		ids := []string{
			string(pushToken.Token),
		}

		c := fcm.NewFcmClient(firebaseApiKey)
		c.NewFcmRegIdsMsg(ids, data)

		status, err := c.Send()

		if err == nil {
			status.PrintResults()
		} else {
			fmt.Println(err)
		}
	}

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
