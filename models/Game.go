package models

import (
	"errors"
	"math/rand"
	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"gopkg.in/asaskevich/govalidator.v4"
)

var (
	GameStateCompleted = 0
	GameStatePending   = 1
	GameStateAborted   = 2
	GameStateStarted   = 3
)

//GameTypes are game modes
var GameTypes = []string{
	"time",
	"points",
}

//Game handels the main part
type Game struct {
	BaseModel

	CreatorRefer  uint `json:"creatorId" valid:"required"`
	OpponentRefer uint `json:"opponentId"`

	Creator  User `json:"creator"`
	Opponent User `json:"opponent"`

	LostRefer uint `json:"lostId"`
	WonRefer  uint `json:"wonId"`

	StateCreator  int `json:"stateCreator"`
	StateOpponent int `json:"stateOpponent"`

	ScoreCreator  int `json:"scoreCreator"`
	ScoreOpponent int `json:"scoreOpponent"`

	StartTimeCreator  *time.Time `json:"startTimeCreator"`
	StartTimeOpponent *time.Time `json:"startTimeOpponent"`

	FromFriendRequest         bool       `json:"fromFriendRequest"`
	FriendRequestAccepted     bool       `json:"accepted"`
	FriendRequestAcceptedTime *time.Time `json:"friendRequestTime"`

	Type  string `json:"type" valid:"required"`
	MapID int    `json:"mapId" valid:"required"`

	LevelRefer uint `json:"levelId" valid:"required"`

	Completed        bool   `json:"completed"`
	AutoCompleted    bool   `json:"autoCompleted"`
	ExtraStringField string `json:"-"`
}

// Save game
func (game *Game) Save() error {
	// validate login request
	_, err := govalidator.ValidateStruct(game)
	if err != nil {
		return err
	}

	db := GetDatabaseSession()
	return db.Save(&game).Error
}

// FindByID game by id
func (game *Game) FindByID(gameID interface{}) error {
	db := GetDatabaseSession()
	result := db.Find(&game, gameID)
	var opponentRefer uint
	rows, _ := db.Raw("SELECT opponent_refer FROM games WHERE id = ?", gameID).Rows()
	for rows.Next() {
		rows.Scan(&opponentRefer)
	}
	game.OpponentRefer = opponentRefer
	fmt.Println("opponentRefer", opponentRefer)
	if result.Error != nil {
		return result.Error
	}
	result = db.Find(&game.Creator, game.CreatorRefer)
	if result.Error != nil {
		return result.Error
	}
	if game.OpponentRefer != 0 {
		result = db.Find(&game.Opponent, game.OpponentRefer)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

// FindMatch finds an already open game
func (game *Game) FindMatch(currentUserID interface{}) error {
	db := GetDatabaseSession()
	sqlQuery := "`opponent_refer` = 0 AND `completed` = 0 AND `state_creator` = ? AND `creator_refer` != ?"
	if res := db.Where(sqlQuery, GameStateCompleted, currentUserID).Order("created_at asc").First(&game); res.Error != nil {
		return res.Error
	}

	// matching game found
	if game.ID == 0 {
		return errors.New("no_mathing_game")
	}
	return nil
}

// FindMatch finds an already open game
func (game Game) FindMatchNew(currentUserID uint) (Game, error) {
	logrus.Info("FindMatchNew")

	db := GetDatabaseSession()
	sqlQuery := "`opponent_refer` = 0 AND `completed` = 0 AND `state_creator` = ? AND `creator_refer` != ?"

	var games []Game
	if res := db.Where(sqlQuery, GameStateCompleted, currentUserID).Order("created_at asc").Find(&games); res.Error != nil {
		return Game{}, res.Error
	}

	var currentUser User
	currentUser.FindByID(currentUserID)
	if currentUser.ID == 0 {
		return Game{}, errors.New("user not found")
	}

	var lastGame Game
	lastGameQuery := "(`creator_refer` = ? OR `opponent_refer` = ?) AND (`creator_refer` != 0 OR `opponent_refer` != 0) AND from_friend_request = 0 AND completed = 1"
	if res := db.Debug().Where(lastGameQuery, currentUserID, currentUserID).Order("updated_at DESC").First(&lastGame); res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return Game{}, res.Error
		}
	}
	fmt.Println(lastGame.ID)

	var matchedGames []Game

	for _, foundGame := range games {
		if lastGame.ID != 0 {
			//check if last game was against or from the same opponent
			if (lastGame.CreatorRefer == foundGame.CreatorRefer) || (lastGame.OpponentRefer == foundGame.CreatorRefer) {
				logrus.Infof("AvoidSameOpponentCheckFailed %d==%d / %v", lastGame.ID, foundGame.ID, lastGame.CreatorRefer == foundGame.CreatorRefer)
				continue
			}
			if (lastGame.CreatorRefer == foundGame.OpponentRefer) || (lastGame.OpponentRefer == foundGame.OpponentRefer) {
				logrus.Infof("AvoidSameOpponentCheckFailed2 %d==%d / %v", lastGame.ID, foundGame.ID, lastGame.CreatorRefer == foundGame.OpponentRefer)
				continue
			}
		}

		var gameCreator User
		gameCreator.FindByID(foundGame.CreatorRefer)
		if gameCreator.ID == 0 {
			return Game{}, errors.New("creator not found")
		}

		var creatorLevel Level
		creatorLevel.FindByID(foundGame.LevelRefer)

		var currentUserLevel Level
		currentUserLevel.FindByID(currentUser.LevelRefer)

		//check if level is one up or one down
		levelDiff := currentUserLevel.Order - creatorLevel.Order
		if levelDiff != 1 && levelDiff != -1 && levelDiff != 0 {
			logrus.Infof("AvoidTooStrongWeakCheck1 %d diff: %d", foundGame.ID, levelDiff)
			continue
		}

		matchedGames = append(matchedGames, foundGame)
	}

	if len(matchedGames) < 1 {
		return Game{}, errors.New("no_mathing_game")
	}

	matchedGame := matchedGames[0]
	return matchedGame, nil
}

//RewardPoints and winner
func (game *Game) RewardPoints() error {
	// General score config
	var baseScoreAddition int = 20
	var baseScoreSubtraction int = -10

	// Add game to the player statistics
	game.Creator.GamesPlayedCount++
	game.Opponent.GamesPlayedCount++

	// Evaluate the winner and update the game and player properties
	var hasCreatorWon bool = game.ScoreCreator < game.ScoreOpponent
	var isDraw bool = game.ScoreCreator == game.ScoreOpponent

	if hasCreatorWon {
		game.WonRefer = game.Creator.ID
		game.LostRefer = game.Opponent.ID
		game.Creator.GamesWonCount++

		game.Creator.Score += baseScoreAddition
		game.Opponent.Score += baseScoreSubtraction

		var pushNotification PushNotification
		go pushNotification.SendGameWonPush(game.Creator)
	} else if !hasCreatorWon && !isDraw {
		game.WonRefer = game.OpponentRefer
		game.LostRefer = game.CreatorRefer
		game.Opponent.GamesWonCount++

		game.Creator.Score += baseScoreSubtraction
		game.Opponent.Score += baseScoreAddition

		var pushNotification PushNotification
		go pushNotification.SendGameLostPush(game.Creator)
	}

	// Check if a player would have a negative value due to the new calculations
	if game.Creator.Score < 0 {
		game.Creator.Score = 0
	}
	if game.Opponent.Score < 0 {
		game.Opponent.Score = 0
	}

	if !hasCreatorWon && game.StateCreator == GameStateAborted {
		game.Creator.Score += (baseScoreAddition / 2)
	}
	if hasCreatorWon && game.StateOpponent == GameStateAborted {
		game.Opponent.Score += (baseScoreAddition / 2)
	}

	// Save users and scores
	if err := game.Creator.Save(); err != nil {
		return err
	}
	if err := game.Opponent.Save(); err != nil {
		return err
	}

	return nil
}

// prepareAndComplete game and reward points
func (game *Game) prepareAndComplete() error {
	if err := game.Creator.FindByID(game.CreatorRefer); err != nil {
		return err
	}

	if err := game.Opponent.FindByID(game.OpponentRefer); err != nil {
		return err
	}

	game.StateCreator = GameStateCompleted
	game.StateOpponent = GameStateCompleted

	if err := game.RewardPoints(); err != nil {
		return err
	}

	game.Completed = true
	game.AutoCompleted = true
	if err := game.Save(); err != nil {
		return err
	}

	return nil
}

// markGameAsLost and subtract points
func (game *Game) markGameAsLost(creatorLost bool) error {
	var loosingPlayer User
	if creatorLost {
		if err := loosingPlayer.FindByID(game.CreatorRefer); err != nil {
			return err
		}
		game.StateCreator = GameStateCompleted
	} else {
		if err := loosingPlayer.FindByID(game.OpponentRefer); err != nil {
			return err
		}
		game.StateOpponent = GameStateCompleted
	}

	newScore := loosingPlayer.Score - 10
	if newScore < 0 {
		newScore = 0
	}
	loosingPlayer.Score = newScore
	loosingPlayer.GamesPlayedCount++

	game.LostRefer = loosingPlayer.ID

	game.Completed = true
	game.AutoCompleted = true

	if err := game.Save(); err != nil {
		return err
	}
	if err := loosingPlayer.Save(); err != nil {
		return err
	}

	return nil
}

// Complete game and reward points
func (game *Game) Complete() error {
	if (game.OpponentRefer == 0) || (game.StateCreator != GameStateCompleted) || (game.StateOpponent != GameStateCompleted) {
		return errors.New("game_not_completed")
	}

	return game.prepareAndComplete()
}

// ForceComplete game and reward points
func (game *Game) ForceComplete() error {
	return game.prepareAndComplete()
}

func (game Game) UnfriendCleanUp(unfrienderID, unfriendedID interface{}) {
	db := GetDatabaseSession()
	sqlQuery := "((creator_refer = ? OR opponent_refer = ?) AND (creator_refer = ? OR opponent_refer = ?)) AND completed != 1"
	db.Unscoped().Where(sqlQuery, unfrienderID, unfriendedID).Delete(&Game{})
	return
}

// CleanUp cleans up open games and aborted ones
func (game Game) CleanUp() {
	db := GetDatabaseSession()

	// Set "global" now for faster date calculations
	now := time.Now()

	// Check for open but unused games
	unansweredGamesDeadline := now.Add(-24 * time.Hour)
	unansweredGamesDeadlineFrom := now.Add(-72 * time.Hour)

	unansweredGamesQuery := "(state_creator = ? AND state_opponent = ?) AND created_at <= ? AND created_at >= ?"

	db.Where(unansweredGamesQuery, GameStateCompleted, GameStatePending, unansweredGamesDeadline, unansweredGamesDeadlineFrom).Delete(&Game{})

	// Check for aborted games
	abortedGamesDeadline := now.Add(-10 * time.Minute)

	abortedGamesCreatorQuery := "(state_creator = ? AND start_time_creator <= ?) AND completed != 1"

	var abortedGamesCreator []Game
	db.Where(abortedGamesCreatorQuery, GameStateStarted, abortedGamesDeadline).Find(&abortedGamesCreator)

	for _, abortedGameCreator := range abortedGamesCreator {
		if abortedGameCreator.CreatorRefer != 0 && abortedGameCreator.OpponentRefer != 0 {
			abortedGameCreator.ScoreCreator = 2
			abortedGameCreator.ScoreOpponent = 1
			abortedGameCreator.ForceComplete()
			return
		}
		abortedGameCreator.markGameAsLost(true)
	}

	abortedGamesOpponentQuery := "(state_opponent = ? AND start_time_opponent <= ?) AND completed != 1"

	var abortedGamesOpponent []Game
	db.Where(abortedGamesOpponentQuery, GameStateStarted, abortedGamesDeadline).Find(&abortedGamesOpponent)

	for _, abortedGameOpponent := range abortedGamesOpponent {
		if abortedGameOpponent.CreatorRefer != 0 && abortedGameOpponent.OpponentRefer != 0 {
			abortedGameOpponent.ScoreCreator = 1
			abortedGameOpponent.ScoreOpponent = 2
			abortedGameOpponent.ForceComplete()
			return
		}
		abortedGameOpponent.markGameAsLost(false)
	}

	return
}

// GameIsAlreadyOpen checks if there is a game open or not
func (game *Game) GameIsAlreadyOpen(userID, friendID interface{}) bool {
	db := GetDatabaseSession()
	sqlQuery := "((creator_refer = ? OR opponent_refer = ?) AND (creator_refer = ? OR opponent_refer = ?)) AND completed != 1"

	var count int
	db.Model(&game).Where(sqlQuery, userID, userID, friendID, friendID).Count(&count)
	return count > 0
}

//SetRandomGameType sets game type
func (game *Game) SetRandomGameType() {
	rand.Seed(time.Now().UTC().UnixNano())
	gameType := GameTypes[rand.Intn(len(GameTypes))]
	game.Type = gameType
}

//SetRandomMapID sets map id
func (game *Game) SetRandomMapID() {
	rand.Seed(time.Now().UTC().UnixNano())
	game.MapID = rand.Intn(20)
}
