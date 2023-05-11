package util

import (
	"encoding/binary"
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

func GenerateId() string {
	rand.Seed(time.Now().Unix())
	return fmt.Sprintf("%d", rand.Int())
}

func EncodeStruct(data any) []byte {
	jsonByte, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error parsing struct - ", err)
		return nil
	}
	return jsonByte
}
func DecodeStruct(source []byte, destination any) error {
	err := json.Unmarshal(source, destination)
	if err != nil {
		return err
	}
	return nil
}

func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
