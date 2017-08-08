package models

import (
	"errors"
	"fmt"
	"time"
	"timedrop/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var err error
var databaseSession *gorm.DB

var (
	ErrUsernameNotUnique = errors.New("username already taken")
	ErrEmailNotUnique    = errors.New("email already in use")
)

//BaseModel to unify model definitions
type BaseModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" sql:"index"`
}

//Bootstrap 's the tables
func Bootstrap() {
	initDatabaseSession()
	db := GetDatabaseSession()
	db.AutoMigrate(&FriendRequest{})
	db.AutoMigrate(&Friend{})
	db.AutoMigrate(&PushToken{})
	db.AutoMigrate(&AuthToken{})
	db.AutoMigrate(&LoginCode{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Game{})
	db.AutoMigrate(&Level{})

	var level Level
	level.Bootstrap()
}

func initDatabaseSession() {
	//databaseSession, err := sql.Open("mysql", config.Cfg.DatabaseSettings.DatabaseUsername + ":"+config.Cfg.DatabaseSettings.DatabasePassword+"@/" + config.Cfg.DatabaseSettings.DatabaseName)
	databaseSession, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
		config.Cfg.DatabaseSettings.DatabaseUsername,
		config.Cfg.DatabaseSettings.DatabasePassword,
		config.Cfg.DatabaseSettings.DatabaseName,
	))
	databaseSession.DB().SetMaxOpenConns(10)
	databaseSession.DB().SetMaxIdleConns(10)
}

//GetDatabaseSession returns (and creates) a session to use
func GetDatabaseSession() *gorm.DB {
	if databaseSession.DB().Ping() != nil {
		initDatabaseSession()
	}
	return databaseSession
}
