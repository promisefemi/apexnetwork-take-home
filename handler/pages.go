package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/promisefemi/apexnetwork-take-home/model"
	"github.com/promisefemi/apexnetwork-take-home/util"
	"log"
	"net/http"
	"time"
)

// Constants
const (
	UserBucket        string = "users"
	TransactionBucket string = "transactions"
	GameSessionBucket string = "gameSession"
	RollSessionBucket string = "rollSession"

	GameStartCost    int = 20
	FirstRowCost     int = 5
	WinningAmount    int = 20
	FundWalletAmount int = 155
)

// ERRORS
var (
	ErrUserNotExist            error = errors.New("user does not exist, kindly create user account ")
	ErrUnableToFundWallet      error = errors.New("unable to fund your wallet, please contact support")
	ErrNoTransactionsAvailable error = errors.New("no transactions available")
	ErrUnableToStartGame       error = errors.New("unable to start game, please contact support")
	ErrGameInSession           error = errors.New("you already have an active game in progress, end previous game to start another ")
	ErrNoGameInSession         error = errors.New("you have no active game in progress, please start a new game ")
	ErrUnableToRollDice        error = errors.New("unable to roll dice, please contact support ")
	ErrNoActiveRollSession     error = errors.New("no roll session active ")
	ErrUnableToEndGame         error = errors.New("unable to end game, please contact support ")
)

// New Handler
type PageHandler struct {
	db *bolt.DB
}

// Create and returns new handler, injects boltDB instance
func NewPageHandler(db *bolt.DB) *PageHandler {
	return &PageHandler{db}
}

// Register new User
func (p *PageHandler) Register(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, "unable to parse form", http.StatusInternalServerError)
	}

	response := model.ApiResponse{
		Status: false,
	}
	//TODO: CHECK why ony x-www-enconde functional

	firstName := r.PostFormValue("first_name")
	lastName := r.PostFormValue("last_name")
	if firstName == "" || lastName == "" {
		response.Message = "Please complete both first and last name"
		p.JSON(response, rw)
		return
	}

	userID := util.GenerateUserId(firstName, lastName)
	user := &model.User{
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		Wallet:    0,
		Asset:     "sat",
	}

	err = p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(UserBucket))
		if err != nil {
			log.Printf("database error - %s", err)
			return fmt.Errorf("error unable to encode struct")
		}
		userByte := util.EncodeStruct(user)
		if userByte == nil {
			return fmt.Errorf("error unable to encode struct")
		}
		return bucket.Put([]byte(userID), userByte)
	})

	if err != nil {
		log.Printf("%s", err)
		response.Message = "Something went wrong, unable to create new user"
		p.JSON(response, rw)
	}

	response.Status = true
	response.Message = "New user created, you can now start games"
	response.Data = user

	p.JSON(response, rw)
	return
}

// Start New Game
func (p *PageHandler) StartGame(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.PostFormValue("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}

	//Validate user
	user, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	//Check if user has funds to start new game
	if user.Wallet < 20 {
		response.Message = "You do not have enough funds to start the game, please fund your account"
		p.JSON(response, rw)
		return
	}

	//Check if there is a game in session
	_, err = p.getActiveGame(userID)
	//if there is an active game, err will be nil
	//return with error (There is an active game)
	if err == nil {
		response.Message = ErrGameInSession.Error()
		p.JSON(response, rw)
		return
	}

	//create new game session
	session := model.GameSession{
		SessionID:  util.GenerateId(),
		UserId:     userID,
		GameStatus: model.INPROGRESS,
	}
	//Create transaction for new game session
	transaction := model.Transaction{
		Type:        model.DEBIT,
		Time:        time.Now().Unix(),
		Description: "Started new Game",
		Amount:      GameStartCost,
		UserID:      userID,
	}
	//Substract new game amount from wallet balance
	user.Wallet -= transaction.Amount

	err = p.db.Update(func(tx *bolt.Tx) error {
		//Get user bucket from bolt db
		userBucket := tx.Bucket([]byte(UserBucket))
		if userBucket == nil {
			log.Printf("database error - %s", err)
			return ErrUserNotExist
		}
		//Get transaction bucket from bolt db
		transactionBucket, err := tx.CreateBucketIfNotExists([]byte(TransactionBucket))
		if err != nil {
			log.Printf("error unable to create transactions bucket - %s", err)
			return ErrUnableToStartGame
		}
		//Get game session bucket from bolt db
		gameSessionBucket, err := tx.CreateBucketIfNotExists([]byte(GameSessionBucket))
		if err != nil {
			log.Printf("error unable to create game session bucket - %s", err)
			return ErrUnableToStartGame
		}

		/*
			. Crete new sequence for transaction
			. Encode Data (transaction, user, gameSession)
			. Update/Insert data into respective buckets
		*/
		id, _ := transactionBucket.NextSequence()
		transactionByte := util.EncodeStruct(transaction)
		if transactionByte == nil {
			log.Println("error encoding transaction struct")
			return ErrUnableToFundWallet
		}

		gameSessionByte := util.EncodeStruct(session)
		if gameSessionByte == nil {
			log.Println("error encoding transaction struct")
			return ErrUnableToFundWallet
		}
		userByte := util.EncodeStruct(user)
		if userByte == nil {
			log.Println("error encoding user struct")
			return ErrUnableToFundWallet
		}

		if err := transactionBucket.Put(util.Itob(int(id)), transactionByte); err != nil {
			log.Println("error inserting transaction")
			return ErrUnableToStartGame
		}
		if err := userBucket.Put([]byte(userID), userByte); err != nil {
			log.Println("error updating user")
			return ErrUnableToStartGame
		}
		if err := gameSessionBucket.Put([]byte(session.SessionID), gameSessionByte); err != nil {
			log.Println("error inserting session")
			return ErrUnableToStartGame
		}
		return nil
	})

	if err != nil {
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	response.Status = true
	response.Message = "Congrats your game session started, you can now roll"
	response.Data = session

	p.JSON(response, rw)
	return
}

// ROLL Dice
func (p *PageHandler) Roll(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.PostFormValue("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate user id
	user, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}
	//Check for active game session
	activeGameSession, err := p.getActiveGame(userID)
	if err != nil {
		response.Message = ErrNoGameInSession.Error()
		p.JSON(response, rw)
		return
	}

	var activeRollSession model.RollSession
	//Check for an active dice roll
	err = p.db.View(func(tx *bolt.Tx) error {
		rollBucket := tx.Bucket([]byte(RollSessionBucket))
		if rollBucket == nil {
			log.Println("error unable to find roll bucket")
			return ErrNoActiveRollSession
		}
		c := rollBucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var rollSession model.RollSession
			if err := util.DecodeStruct(v, &rollSession); err != nil {
				log.Printf("unable to parse session - %s", err)
				continue
			}
			if rollSession.GameSessionID == activeGameSession.SessionID && rollSession.RowStatus == model.INPROGRESS {
				activeRollSession = rollSession
				return nil
			}
		}
		return ErrNoActiveRollSession
	})

	//If there is an err, that means there is no active roll session
	//So roll first dice
	if err != nil && err == ErrNoActiveRollSession {
		//Check if wallet balance is enough to row first dice
		if user.Wallet < FirstRowCost {
			response.Message = "You do not have enough funds to roll dice, please fund your account"
			p.JSON(response, rw)
			return
		}

		//Structuring Data models
		newRoll := model.RollSession{
			GameSessionID: activeGameSession.SessionID,
			WinningGame:   util.GenerateDiceSessionRoll(),
			FirstRoll:     util.GenerateDiceRoll(),
			RowStatus:     model.INPROGRESS,
			UserID:        userID,
			RollID:        util.GenerateId(),
		}

		transaction := model.Transaction{
			Type:        model.DEBIT,
			Time:        time.Now().Unix(),
			Description: "Rolled dice",
			Amount:      FirstRowCost,
			UserID:      userID,
		}

		user.Wallet -= transaction.Amount

		/*
			. Get Respective Buckets (User, Transaction, Dice Roll)
			. Encode data
			. Update/Insert data into respective buckets
		*/
		err = p.db.Update(func(tx *bolt.Tx) error {
			userBucket := tx.Bucket([]byte(UserBucket))
			if userBucket == nil {
				log.Printf("database error - %s", err)
				return ErrUnableToRollDice
			}

			transactionBucket, err := tx.CreateBucketIfNotExists([]byte(TransactionBucket))
			if err != nil {
				log.Printf("error unable to create transactions bucket - %s", err)
				return ErrUnableToRollDice
			}

			rollSessionBucket, err := tx.CreateBucketIfNotExists([]byte(RollSessionBucket))
			if err != nil {
				log.Printf("error unable to create roll session bucket - %s", err)
				return ErrUnableToRollDice
			}

			id, _ := transactionBucket.NextSequence()
			transactionByte := util.EncodeStruct(transaction)
			if transactionByte == nil {
				log.Println("error encoding transaction struct")
				return ErrUnableToRollDice
			}

			gameSessionByte := util.EncodeStruct(newRoll)
			if gameSessionByte == nil {
				log.Println("error encoding transaction struct")
				return ErrUnableToRollDice
			}
			userByte := util.EncodeStruct(user)
			if userByte == nil {
				log.Println("error encoding user struct")
				return ErrUnableToRollDice
			}

			if err := transactionBucket.Put(util.Itob(int(id)), transactionByte); err != nil {
				log.Println("error inserting transaction")
				return ErrUnableToRollDice
			}
			if err := userBucket.Put([]byte(userID), userByte); err != nil {
				log.Println("error inserting user")
				return ErrUnableToRollDice
			}
			if err := rollSessionBucket.Put([]byte(newRoll.RollID), gameSessionByte); err != nil {
				log.Println("error inserting session")
				return ErrUnableToRollDice
			}
			return nil
		})

		//Handle error
		if err != nil {
			response.Message = err.Error()
			p.JSON(response, rw)
			return
		}
		//Return the number rolled and how many they have to roll to win
		response.Message = fmt.Sprintf("Congrats, you rolled %d to win you have to to roll %d 🤞", newRoll.FirstRoll, newRoll.WinningGame-newRoll.FirstRoll)
		response.Status = true
		p.JSON(response, rw)
		return

	} else {
		//If there is an active Roll
		//Generate random dice roll for second roll
		//Set the active Dice Roll as completed
		rollNewDice := util.GenerateDiceRoll()
		activeRollSession.SecondRoll = rollNewDice
		activeRollSession.RowStatus = model.COMPLETED
		/*
			. Get data respective bucket
			. Encode dice roll
			. Update active dice roll
		*/
		err := p.db.Update(func(tx *bolt.Tx) error {
			rollSessionBucket := tx.Bucket([]byte(RollSessionBucket))
			if rollSessionBucket == nil {
				log.Printf("error unable to create roll session bucket - %s", err)
				return ErrUnableToStartGame
			}
			gameSessionByte := util.EncodeStruct(activeRollSession)
			if gameSessionByte == nil {
				log.Println("error encoding dice roll struct")
				return ErrUnableToRollDice
			}

			if err := rollSessionBucket.Put([]byte(activeRollSession.RollID), gameSessionByte); err != nil {
				log.Println("error inserting session")
				return ErrUnableToRollDice
			}
			return nil
		})

		//Handle Error
		if err != nil {
			response.Message = err.Error()
			p.JSON(response, rw)
			return
		}

		//Check for winnings
		if (activeRollSession.FirstRoll + activeRollSession.SecondRoll) == activeRollSession.WinningGame {
			/*
				. If user won dice roll
				. Create transaction data for wallet credit
				. Update user Wallet
			*/
			transaction := model.Transaction{
				Type:        model.CREDIT,
				Time:        time.Now().Unix(),
				Description: "Winnings",
				Amount:      WinningAmount,
				UserID:      userID,
			}

			user.Wallet += transaction.Amount

			err = p.db.Update(func(tx *bolt.Tx) error {
				userBucket := tx.Bucket([]byte(UserBucket))
				if userBucket == nil {
					log.Printf("database error - %s", err)
					return ErrUnableToRollDice
				}

				transactionBucket, err := tx.CreateBucketIfNotExists([]byte(TransactionBucket))
				if err != nil {
					log.Printf("error unable to create transactions bucket - %s", err)
					return ErrUnableToRollDice
				}

				id, _ := transactionBucket.NextSequence()
				transactionByte := util.EncodeStruct(transaction)
				if transactionByte == nil {
					log.Println("error encoding transaction struct")
					return ErrUnableToRollDice
				}

				userByte := util.EncodeStruct(user)
				if userByte == nil {
					log.Println("error encoding user struct")
					return ErrUnableToRollDice
				}

				if err := transactionBucket.Put(util.Itob(int(id)), transactionByte); err != nil {
					log.Println("error inserting transaction")
					return ErrUnableToRollDice
				}
				if err := userBucket.Put([]byte(userID), userByte); err != nil {
					log.Println("error inserting user")
					return ErrUnableToRollDice
				}
				return nil
			})

			if err != nil {
				response.Message = err.Error()
				p.JSON(response, rw)
				return
			}

			response.Status = true
			response.Message = fmt.Sprintf("Hurray 🤑, you have won %d, do you want to try again ", WinningAmount)
			p.JSON(response, rw)
			return

		} else {
			//User did not win, reply with message
			response.Status = true
			response.Message = fmt.Sprintf("Oops 😥, you did not win you rolled %d, but you can try again ", activeRollSession.SecondRoll)
			p.JSON(response, rw)
			return
		}

	}

}

func (p *PageHandler) EndGame(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.PostFormValue("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate user account
	_, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	/*
		. Check for inprogress games
		. Update as completed
		. Check for inprogress dice roll
		. Update as completed
	*/
	err = p.db.Update(func(tx *bolt.Tx) error {
		gameSessionBucket := tx.Bucket([]byte(GameSessionBucket))
		if gameSessionBucket == nil {
			log.Println("unable to get session Bucket")
			return ErrUnableToEndGame
		}

		rollGameBucket := tx.Bucket([]byte(RollSessionBucket))
		if rollGameBucket == nil {
			log.Println("unable to get session Bucket")
			return ErrUnableToEndGame
		}

		gameCursor := gameSessionBucket.Cursor()
		for k, v := gameCursor.First(); k != nil; k, v = gameCursor.Next() {
			var gameSession model.GameSession
			err := util.DecodeStruct(v, &gameSession)
			if err != nil {
				log.Println("unable to decode struct")
				return ErrUnableToEndGame
			}
			if gameSession.UserId == userID && gameSession.GameStatus == model.INPROGRESS {
				gameSession.GameStatus = model.COMPLETED
				gameSessionByte := util.EncodeStruct(gameSession)
				err := gameSessionBucket.Put([]byte(gameSession.SessionID), gameSessionByte)
				if err != nil {
					log.Println("unable to update game session struct")
					return ErrUnableToEndGame
				}
			}
		}

		rollCursor := rollGameBucket.Cursor()
		for k, v := rollCursor.First(); k != nil; k, v = rollCursor.Next() {
			var rollSession model.RollSession
			err := util.DecodeStruct(v, &rollSession)
			if err != nil {
				log.Println("unable to decode struct")
				return ErrUnableToEndGame
			}
			if rollSession.UserID == userID && rollSession.RowStatus == model.INPROGRESS {
				rollSession.RowStatus = model.COMPLETED
				rollSessionByte := util.EncodeStruct(rollSession)
				err := gameSessionBucket.Put([]byte(rollSession.RollID), rollSessionByte)
				if err != nil {
					log.Println("unable to update roll session struct")
					return ErrUnableToEndGame
				}
			}
		}
		return nil
	})

	//Handle Error
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	response.Status = true
	response.Message = "Successfully ended all game, we hope to see you again"
	p.JSON(response, rw)
	return
}

func (p *PageHandler) CheckActiveGame(rw http.ResponseWriter, r *http.Request) {

	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate user
	_, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}
	//Check for active Game
	activeGameSession, err := p.getActiveGame(userID)
	//Handle error
	if err != nil {
		response.Message = ErrNoGameInSession.Error()
		p.JSON(response, rw)
		return
	}

	response.Status = true
	response.Data = activeGameSession
	p.JSON(response, rw)
	return
}

func (p *PageHandler) FundWallet(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.PostFormValue("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate user
	user, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}
	//Check if wallet balance is small enough to allow funding
	if user.Wallet > 35 {
		response.Message = "Unfortunately you cannot fund your wallet unless its less than 35"
		response.Data = user
		p.JSON(response, rw)
		return
	}
	/*
		. Create structure for traansaction
		. Update user wallet
	*/
	transaction := model.Transaction{
		Type:        model.CREDIT,
		Time:        time.Now().Unix(),
		Description: "Wallet Funding",
		Amount:      FundWalletAmount,
		UserID:      userID,
	}

	user.Wallet += transaction.Amount

	/*
		. Get bucket for respective data
		. Encode data
		. Insert/Update data in respective bucket
	*/
	err = p.db.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(UserBucket))
		if userBucket == nil {
			log.Printf("database error - %s", err)
			return ErrUserNotExist
		}

		transactionBucket, err := tx.CreateBucketIfNotExists([]byte(TransactionBucket))
		if err != nil {
			log.Printf("error unable to create transactions bucket - %s", err)
			return ErrUnableToFundWallet
		}

		id, _ := transactionBucket.NextSequence()
		transactionByte := util.EncodeStruct(transaction)
		if transactionByte == nil {
			log.Println("error encoding transaction struct")
			return ErrUnableToFundWallet
		}
		userByte := util.EncodeStruct(user)
		if userByte == nil {
			log.Println("error encoding user struct")
			return ErrUnableToFundWallet
		}

		if err := transactionBucket.Put(util.Itob(int(id)), transactionByte); err != nil {
			log.Println("error inserting transaction")
			return ErrUnableToFundWallet
		}
		if err := userBucket.Put([]byte(userID), userByte); err != nil {
			log.Println("error inserting transaction")
			return ErrUnableToFundWallet
		}
		return nil
	})
	//Handle Error
	if err != nil {
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	response.Status = true
	response.Message = "Wallet funding successful"
	response.Data = user

	p.JSON(response, rw)
	return
}

func (p *PageHandler) GetWalletBalance(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate User
	user, err := p.getUser(userID)
	//Handle Error
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}
	//Return user (user detail contains wallet)
	response.Status = true
	response.Data = user
	p.JSON(response, rw)
	return

}

func (p *PageHandler) Transactions(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	response := model.ApiResponse{
		Status: false,
	}
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		response.Message = "Please enter User ID"
		p.JSON(response, rw)
	}
	//Validate user details
	_, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	transactions := make([]model.Transaction, 0)
	/*
		. Get transaction bucket
		. Go through each of them
		. append into transactions slice
	*/
	err = p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(TransactionBucket))
		if bucket == nil {
			log.Printf("error no transactions table - %s", err)
			return ErrNoTransactionsAvailable
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var transaction model.Transaction
			err := util.DecodeStruct(v, &transaction)
			if err != nil {
				log.Printf("error decoding byte to struct %s", err)
				continue
			}
			if transaction.UserID == userID {
				transactions = append(transactions, transaction)
			}
		}
		return nil
	})

	//Handle error
	if err != nil {
		response.Message = err.Error()
		response.Data = transactions
		p.JSON(response, rw)
		return
	}

	response.Status = true
	response.Data = transactions
	p.JSON(response, rw)
	return
}

func (p *PageHandler) getActiveGame(userID string) (*model.GameSession, error) {
	var activeSession model.GameSession

	err := p.db.View(func(tx *bolt.Tx) error {
		gameSessionBucket := tx.Bucket([]byte(GameSessionBucket))
		if gameSessionBucket == nil {
			log.Println("error unable to get game session bucket")
			return errors.New("unable to get game bucket")
		}

		c := gameSessionBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := util.DecodeStruct(v, &activeSession)
			if err != nil {
				log.Printf("error unable to parse game session - %s", err)
				continue
			}
			if activeSession.UserId == userID && activeSession.GameStatus == model.INPROGRESS {
				return nil
			}
		}
		return ErrNoGameInSession
	})

	if err != nil {
		return nil, err
	}

	return &activeSession, nil

}
func (p *PageHandler) getUser(userID string) (*model.User, error) {
	var user model.User
	err := p.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(UserBucket))
		if bucket == nil {
			log.Println("error getting db bucket ")
			return ErrUserNotExist
		}

		userByte := bucket.Get([]byte(userID))
		fmt.Printf("%s\n", userByte)
		err := util.DecodeStruct(userByte, &user)
		if err != nil {
			log.Printf("error parsing struct - %s", err)
			return ErrUserNotExist
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *PageHandler) JSON(data any, rw http.ResponseWriter) {
	rw.Header().Add("Content-Type", "application/json")
	jsonByte, err := json.Marshal(data)
	if err != nil {
		rw.WriteHeader(500)
		return
	}
	_, err = rw.Write(jsonByte)
}
