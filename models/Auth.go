package models

//LoginCode is used for user login
type LoginCode struct {
	BaseModel

	Code string `gorm:";unique_index"`

	IsVerifyEmail bool   `json:"isVerifyEmail"`
	Email         string `json:"email" valid:"email"`
}

//AuthToken is used for the actual athentification via JWT token
type AuthToken struct {
	BaseModel

	UserRefer uint
	Token     string `gorm:";unique_index"`
}

//FindByUserRefer finds a AuthToken by user id
func (authToken *AuthToken) FindByUserRefer(id interface{}) (err error) {
	db := GetDatabaseSession()
	return db.First(&authToken, id).Error
}
