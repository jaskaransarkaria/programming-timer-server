package utils

import (
	"github.com/google/uuid"
	"log"
	"errors"
)

// RandomGenerator ... IDs for Sessions & Users
type RandomGenerator func(typeOfID string) string

// GenerateRandomID generates session & user ids
func GenerateRandomID(typeOfID string) string {
	length, err := getIDLength(typeOfID)
		if err != nil {
			log.Println("err generating ID", err)
		}
	uuid := uuid.New().String()
	return uuid[:length]
}

func getIDLength(typeOfID string) (int8, error) {
		if (typeOfID == "session") {
		return 4, nil
	}
	if (typeOfID == "user") {
		return 8, nil
	}
	return -1, errors.New("Invalid typeofID")
}
