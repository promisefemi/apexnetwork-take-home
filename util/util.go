package util

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func GenerateUserId(firstname, lastname string) string {
	rand.Seed(time.Now().Unix())
	return fmt.Sprintf("%s-%s-%d", strings.ToLower(firstname), strings.ToLower(lastname), rand.Int())
}

func EncodeStruct(data any) []byte {
	jsonByte, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error parsing struct - ", err)
		return nil
	}
	return jsonByte
}
