package handler

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type PageHandler struct {
	db *bolt.DB
}

func NewPageHandler(db *bolt.DB) *PageHandler {
	return &PageHandler{db}
}

func (p *PageHandler) Register(rw http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	response := &ApiResponse{
		Status: false,
	}

	firstName := r.PostFormValue("first_name")
	lastName := r.PostFormValue("first_name")
	if firstName == "" || lastName == "" {
		response.Message = "Please complete both first and last name"
		p.JSON(response, rw)
		return
	}

	usedId := generateUserId(firstName, lastName)

	p.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			log.Printf("database error - %s", err)
			http.Error(rw, "Something went wrong", http.StatusInternalServerError)
		}

	})

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

func generateUserId(firstname, lastname string) string {
	rand.Seed(time.Now().Unix())
	return fmt.Sprintf("%s-%s-%d", strings.ToLower(firstname), strings.ToLower(lastname), rand.Int())
}
