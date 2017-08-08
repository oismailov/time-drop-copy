package models

// Level 's
type Level struct {
	BaseModel

	FromScore int    `json:"fromScore"`
	ToScore   int    `json:"toScore"`
	Name      string `json:"name" gorm:";unique_index"`
	Order     int    `json:"order" gorm:";unique_index"`
}

//Bootstrap given levels
func (level Level) Bootstrap() {
	db := GetDatabaseSession()
	novice := Level{
		Name:      "Novice",
		FromScore: 0,
		ToScore:   399,
		Order:     1,
	}
	db.Save(&novice)

	greenhorn := Level{
		Name:      "Greenhorn",
		FromScore: novice.ToScore + 1,
		ToScore:   799,
		Order:     novice.Order + 1,
	}
	db.Save(&greenhorn)

	export := Level{
		Name:      "Expert",
		FromScore: greenhorn.ToScore + 1,
		ToScore:   1099,
		Order:     greenhorn.Order + 1,
	}
	db.Save(&export)

	master := Level{
		Name:      "Master",
		FromScore: export.ToScore + 1,
		ToScore:   1399,
		Order:     export.Order + 1,
	}
	db.Save(&master)

	grandMaster := Level{
		Name:      "Grand Master",
		FromScore: master.ToScore + 1,
		ToScore:   1699,
		Order:     master.Order + 1,
	}
	db.Save(&grandMaster)

	legend := Level{
		Name:      "Legend",
		FromScore: grandMaster.ToScore + 1,
		ToScore:   1999,
		Order:     grandMaster.Order + 1,
	}
	db.Save(&legend)

	divine := Level{
		Name:      "Divine",
		FromScore: grandMaster.ToScore + 1,
		ToScore:   2999,
		Order:     legend.Order + 1,
	}
	db.Save(&divine)

	splasher := Level{
		Name:      "Splasher",
		FromScore: grandMaster.ToScore + 1,
		ToScore:   9999999,
		Order:     divine.Order + 1,
	}
	db.Save(&splasher)
}

//FindByID finds a level by id
func (level *Level) FindByID(id interface{}) (err error) {
	db := GetDatabaseSession()
	return db.First(&level, id).Error
}

//FindByScore find level by user score
func (level *Level) FindByScore(score int) {
	db := GetDatabaseSession()
	result := db.Where("from_score <= ? AND to_score >= ?", score, score).First(&level)

	if level.ID == 0 || result.Error != nil {
		db.First(&level)
	}
}

//FindByName find level by name
func (level *Level) FindByName(name string) {
	db := GetDatabaseSession()
	result := db.Where("name = ?", name).First(&level)

	if level.ID == 0 || result.Error != nil {
		db.First(&level)
	}
}
