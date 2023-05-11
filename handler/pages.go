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

const (
	UserBucket        string = "users"
	TransactionBucket string = "transactions"
)

var (
	ErrUserNotExist            error = errors.New("user does not exist, kindly create user account ")
	ErrUnableToFundWallet      error = errors.New("unable to fund your wallet, please contact support")
	ErrNoTransactionsAvailable error = errors.New("no transactions available")
)

type PageHandler struct {
	db *bolt.DB
}

func NewPageHandler(db *bolt.DB) *PageHandler {
	return &PageHandler{db}
}

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

func (p *PageHandler) StartGame(rw http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) Roll(rw http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) EndGame(rw http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) CheckActiveGame(rw http.ResponseWriter, r *http.Request) {

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

	user, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	if user.Wallet > 35 {
		response.Message = "Unfortunately you cannot fund your wallet unless its less than 35"
		response.Data = user
		p.JSON(response, rw)
		return
	}

	transaction := model.Transaction{
		Type:   model.CREDIT,
		Time:   time.Now().Unix(),
		Amount: 155,
		UserID: userID,
	}

	user.Wallet += transaction.Amount

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

	user, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

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

	_, err := p.getUser(userID)
	if err != nil {
		log.Println(err)
		response.Message = err.Error()
		p.JSON(response, rw)
		return
	}

	transactions := make([]model.Transaction, 0)

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
