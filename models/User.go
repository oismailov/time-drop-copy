package models

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"time"

	"timedrop/helpers"

	log "github.com/inconshreveable/log15"
	"gopkg.in/asaskevich/govalidator.v4"
	"gopkg.in/gomail.v2"
)

var UserLanguageDe = "de_DE"
var UserLanguageEn = "en_US"

var userLogger = log.New("models", "user")

//User struct handels user
type User struct {
	BaseModel

	Username string `json:"username" gorm:";unique_index" valid:"required"`
	Email    string `json:"email" valid:"email"`
	Language string `json:"language"`

	Guest bool `json:"guest"`

	FacebookID    string    `json:"facebookId"`
	FbImageUrl    string    `json:"fbImageUrl"`
	Avatar        int       `json:"avatar"`
	FbName        string    `json:"fbName"`
	IsVerified    bool      `json:"isVerified"`
	VerifyCode    string    `json:"verifyCode" gorm:"-"`
	ExtraData     string    `json:"extraData"`
	UserUpdatedAt time.Time `json:"userUpdatedAt"`

	Score            int `json:"score"`
	GamesPlayedCount int `json:"gamesPlayedCount"`
	GamesWonCount    int `json:"gamesWonCount"`

	CurrentLevel int    `json:"currentLevel"`
	Level        string `json:"level"`
	LevelRefer   uint   `json:"levelRefer"`
	LevelData    string `json:"levelData"`

	TopLevel      string `json:"topLevel"`
	TopLevelRefer uint   `json:"topLevelRefer"`

	AchievementsData string `json:"achievementData"`
	TreasureData     string `json:"treasureData"`
	PergamentData    string `json:"pergamentData"`
	Coins            int    `json:"coins"`

	LoginCodes []LoginCode `json:"-" gorm:"many2many:user_logincodes;"`
	PushTokens []PushToken `gorm:"ForeignKey:UserRefer" json:"-"`
	AuthTokens []AuthToken `json:"-" gorm:"ForeignKey:UserRefer"`
}

type UserLoginToken struct {
	Token string `gorm:code`
}

//FindByID finds a user with id
func (user *User) FindByID(id interface{}) (err error) {
	db := GetDatabaseSession()
	return db.First(&user, id).Error
}

//FindByUsername finds a user by username
func (user *User) FindByUsername(username string) (err error) {
	db := GetDatabaseSession()
	return db.Where("username = ?", username).First(&user).Error
}

//FindByEmail finds a user by email
func (user *User) FindByEmail(email string) (err error) {
	db := GetDatabaseSession()
	return db.Where("email = ?", email).First(&user).Error
}

//FindByFacebookID finds a user by facebookID
func (user *User) FindByFacebookID(id string) (err error) {
	db := GetDatabaseSession()
	return db.Where("facebook_id = ?", id).First(&user).Error
}

//FindOtherUserByFacebookIDAndID finds a user by facebookID and ID
func (user *User) FindOtherUserByFacebookIDAndID(id uint, facebookID string) (err error) {
	db := GetDatabaseSession()
	return db.Where("facebook_id = ? AND id != ?", facebookID, id).First(&user).Error
}

//FindByToken finds a user by jwt token
func (user *User) FindByToken(token string) (err error) {
	db := GetDatabaseSession()
	var authToken AuthToken
	result := db.Where("token = ?", token).Find(&authToken)
	if result.Error != nil {
		return result.Error
	}
	if authToken.Token != token {
		return errors.New("token_not_found")
	}

	err = user.FindByID(authToken.UserRefer)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("token_user_not_found")
	}

	return nil
}

func (user *User) DoesTokenExist() bool {
	db := GetDatabaseSession()
	var count int
	db.Where("user_refer = ?", user.ID).Count(&count)
	if count == 0 {
		return false
	}

	return true
}

//Save or create user
func (user *User) Save() error {
	db := GetDatabaseSession()

	if user.Language == "de" || user.Language == UserLanguageDe {
		user.Language = UserLanguageDe
	} else {
		user.Language = UserLanguageEn
	}

	var level Level
	level.FindByScore(user.Score)
	user.Level = level.Name
	user.LevelRefer = level.ID

	if user.TopLevel == "" {
		user.TopLevel = user.Level
		user.TopLevelRefer = user.LevelRefer
	} else {
		var topLevel Level
		topLevel.FindByID(user.TopLevelRefer)
		if topLevel.ID != 0 {
			if level.Order > topLevel.Order {
				user.TopLevel = user.Level
				user.TopLevelRefer = user.LevelRefer
			}
		}
	}

	user.UserUpdatedAt = time.Now()

	result := db.Save(&user)
	return result.Error
}

//AppendLoginCode to user obj
func (user *User) AppendLoginCode(loginCode LoginCode) {
	if user.LoginCodes == nil {
		var loginCodes []LoginCode
		user.LoginCodes = loginCodes
	}
	user.LoginCodes = append(user.LoginCodes, loginCode)
}

//AppendAuthToken to user obj
func (user *User) AppendAuthToken(authTokens AuthToken) {
	if user.AuthTokens == nil {
		var authTokens []AuthToken
		user.AuthTokens = authTokens
	}
	user.AuthTokens = append(user.AuthTokens, authTokens)
}

//IsEmailUnique checks if email is unique
func (user User) IsEmailUnique(email string) bool {
	db := GetDatabaseSession()

	var count int
	db.Model(&User{}).Where("email = ?", email).Count(&count)
	return count == 0
}

//SendLoginEmail to the current user
func (user *User) sendTokenEmail(loginCode LoginCode, destinationEmail string) error {
	if _, err := govalidator.ValidateStruct(user); err != nil {
		return err
	}

	if destinationEmail == "" {
		return errors.New("no email address found")
	}

	user.AppendLoginCode(loginCode)
	user.Save()

	d := gomail.NewDialer("asmtp.mail.hostpoint.ch", 587, "no-reply@time-drop.com", "TD16$MPT20WP!")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	s, err := d.Dial()
	if err != nil {
		userLogger.Error(fmt.Sprintf("Err while sending email: %v", err))
		return err
	}

	username := user.Username
	if username == "" {
		username = user.Email
	}

	userLogger.Debug(fmt.Sprintf("Try to send email to <%s> %s", user.Email, username))

	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@time-drop.com")
	m.SetAddressHeader("To", destinationEmail, username)
	m.SetHeader("Subject", helpers.TranslateStr("verify_code_email_subject", user.Language))
	m.SetBody("text/plain", fmt.Sprintf(helpers.TranslateStr("verify_code_email_body", user.Language), loginCode.Code))

	go func() {
		if err := gomail.Send(s, m); err != nil {
			userLogger.Error(fmt.Sprintf("Could not send email to %q: %v", user.Email, err))
		}
		m.Reset()
	}()

	return nil
}

//SendLoginEmail to the current user
func (user *User) SendLoginEmail() error {
	if _, err := govalidator.ValidateStruct(user); err != nil {
		fmt.Println(err)
		return err
	}

	authToken := LoginCode{
		Code:          helpers.GenerateOneTimeToken(),
		IsVerifyEmail: false,
	}
	user.sendTokenEmail(authToken, user.Email)

	return nil
}

//SendEmailVerification to the current user
func (user *User) SendEmailVerification(newEmail string) error {
	if _, err := govalidator.ValidateStruct(user); err != nil {
		return err
	}

	authToken := LoginCode{
		Code:          helpers.GenerateOneTimeToken(),
		IsVerifyEmail: true,
		Email:         newEmail,
	}

	if _, err := govalidator.ValidateStruct(authToken); err != nil {
		return err
	}

	return user.sendTokenEmail(authToken, newEmail)
}

func (user *User) SendAuthToken(token string) error {

	if _, err := govalidator.ValidateStruct(user); err != nil {
		return err
	}

	authToken := LoginCode{
		Code:          token,
		IsVerifyEmail: false,
		Email:         user.Email,
	}

	return user.sendTokenEmail(authToken, user.Email)
}

//ValidateLoginCode validates the login code
func (user *User) ValidateLoginCode(token string) bool {
	db := GetDatabaseSession()

	var checkUser User
	db.Preload("LoginCodes", "code = ?", token).First(&checkUser, user.ID)
	if checkUser.LoginCodes == nil || len(checkUser.LoginCodes) > 1 {
		return false
	}

	if result := db.Delete(&checkUser.LoginCodes[0]); result.Error != nil {
		return false
	}

	return true
}

//ValidateUserLoginCode validates the login code
func (user *User) ValidateUserLoginCode(code string, userId uint) bool {
	db := GetDatabaseSession()
	var loginCode UserLoginToken
	db.Raw(`
		SELECT lc.code AS code
		FROM login_codes AS lc
		JOIN user_logincodes AS ulc ON ulc.login_code_id = lc.id
		WHERE lc.code = ?
		AND ulc.user_id = ?`, code, userId).Row().Scan(&loginCode.Token)
	if loginCode.Token != "" {
		return true
	}

	return false
}

//ValidateEmailCode validates the email code
func (user *User) ValidateEmailCode(code string) (email string, err error) {
	tmpLog := userLogger.New("func", "ValidateEmailCode")
	db := GetDatabaseSession()

	var checkUser User
	db.Preload("LoginCodes", "code = ?", code).First(&checkUser, user.ID)
	if checkUser.LoginCodes == nil || len(checkUser.LoginCodes) > 1 {
		return "", errors.New("code_not_found")
	}

	loginCode := checkUser.LoginCodes[0]
	if !loginCode.IsVerifyEmail {
		return "", errors.New("code_not_found")
	}

	if result := db.Delete(&checkUser.LoginCodes[0]); result.Error != nil {
		tmpLog.Error(fmt.Sprintf("code couldn't be invalidated due to %v", result.Error))
		return "", errors.New("code_not_invalidated")
	}

	return loginCode.Email, nil
}

//AddPushToken adds push token
func (user *User) AddPushToken(pushToken PushToken) error {
	tmpLog := userLogger.New("func", "AddPushToken")
	tmpLog.Debug(fmt.Sprintf("added push token '%s' for user '%d'", pushToken.Token, user.ID))
	user.PushTokens = append(user.PushTokens, pushToken)
	return user.Save()
}

//DeletePushToken removes a push token
func (user *User) DeletePushToken(pushToken string) error {
	db := GetDatabaseSession()
	queryToken := PushToken{
		Token: pushToken,
	}
	return db.Unscoped().Where(queryToken).Delete(&PushToken{}).Error
}

//GetAPNSTokens gets the push token
func (user *User) GetAPNSTokens() []PushToken {
	db := GetDatabaseSession()

	var pushTokens []PushToken

	queryToken := PushToken{
		UserRefer: user.ID,
		Platform:  "ios",
	}

	db.Model(&PushToken{}).Where(queryToken).Find(&pushTokens)
	return pushTokens
}

//GetFireBaseTokens gets the push token
func (user *User) GetFireBaseTokens() []PushToken {
	db := GetDatabaseSession()

	var pushTokens []PushToken

	queryToken := PushToken{
		UserRefer: user.ID,
		Platform:  "android",
	}

	db.Model(&PushToken{}).Where(queryToken).Find(&pushTokens)
	return pushTokens
}

//GetGuestUsername returns username for a guest
func GetGuestUsername(guestId int) string {
	db := GetDatabaseSession()

	var user User
	fmt.Printf("guestId: %+v\n", guestId)
	if guestId == 0 {
		var count int
		db.Model(&User{}).Count(&count)
		guestId = count + 1
	}
	userName := "TD" + strconv.Itoa(guestId)
	userQuery := User{
		Username: userName,
	}

	db.Where(userQuery).Find(&user)
	if user.ID != 0 {
		return GetGuestUsername(guestId + 1)
	}

	return userName
}

//FindAllByFacebookIDs finds a user by facebookID
func (user *User) FindAllByFacebookIDs(friends []uint, userId uint) []int {
	db := GetDatabaseSession()
	users := []int{}
	rows, _ := db.Raw("select id from users where facebook_id in (?) and id != ?", friends, userId).Rows()
	defer rows.Close()
	for rows.Next() {
		var user int
		rows.Scan(&user)
		users = append(users, user)
	}

	return users
}

//AddUserFriends
func (user *User) AddUserFriends(users []int, userId uint) {
	db := GetDatabaseSession()
	for _, user := range users {
		db.Exec("DELETE FROM friend_requests WHERE (requester_refer = ? OR receiver_refer = ?) AND (requester_refer = ? OR receiver_refer = ?)",
			userId, userId, user, user)
		db.Exec("INSERT IGNORE INTO friends (requester_refer, receiver_refer, created_at) VALUES (?, ?, NOW())", userId, user)
	}
}

func (user *User) DeleteFacebookData() {
	db := GetDatabaseSession()
	db.Exec("UPDATE users SET facebook_id = null, fb_image_url = null, fb_name = null, is_verified = null WHERE id = ? LIMIT 1", user.ID)
}

func (user *User) DeleteEmailData() {
	db := GetDatabaseSession()
	db.Exec("UPDATE users SET email = null, is_verified = null WHERE id = ? LIMIT 1", user.ID)
	db.Exec("DELETE FROM login_codes WHERE id IN (SELECT login_code_id FROM user_logincodes WHERE user_id = ?)", user.ID)
	db.Exec("DELETE FROM user_logincodes WHERE user_id = ?", user.ID)
}

func (user *User) UpdateUserUpdatedAt() {
	db := GetDatabaseSession()
	db.Exec("UPDATE users SET user_updated_at = NOW() WHERE id = ?", user.ID)
}

func (user *User) FindTokenIdByUserId() int {
	db := GetDatabaseSession()
	rows, _ := db.Raw("select id from auth_tokens where user_refer = ?  limit 1", user.ID).Rows()
	defer rows.Close()
	var tokenId int
	for rows.Next() {
		rows.Scan(&tokenId)
	}

	return tokenId
}
