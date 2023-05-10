package handler

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/promisefemi/apexnetwork-take-home/model"
	"github.com/promisefemi/apexnetwork-take-home/util"
	"log"
	"net/http"
)

type PageHandler struct {
	db *bolt.DB
}

func NewPageHandler(db *bolt.DB) *PageHandler {
	return &PageHandler{db}
}

func (p *PageHandler) Register(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	response := &model.ApiResponse{
		Status: false,
	}

	firstName := r.PostFormValue("first_name")
	lastName := r.PostFormValue("first_name")
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

	err := p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("users"))
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

func (p *PageHandler) Start(res http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) Roll(res http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) End(res http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) FundWallet(res http.ResponseWriter, r *http.Request) {

}
func (p *PageHandler) GetWalletBalance(res http.ResponseWriter, r *http.Request) {

}

func (p *PageHandler) JSON(data any, res http.ResponseWriter) {
	res.Header().Add("Content-Type", "application/json")
	jsonByte, err := json.Marshal(data)
	if err != nil {
		res.WriteHeader(500)
		return
	}
	_, err = res.Write(jsonByte)
}
