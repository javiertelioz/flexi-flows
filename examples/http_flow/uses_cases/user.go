package use_cases

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// User represents user information.
type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Address  Address `json:"address"`
}

// Address represents user address.
type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

// GetUserFunc
// @type: task
// @description Retrieve user information by id.
//
// @input userID (int): User ID.
//
// @output User: Returns user information.
// @output error: Returns an error if user no exist.
func GetUserFunc(data any) (any, error) {
	userID, ok := data.(int)
	if !ok {
		return nil, fmt.Errorf("GetUserFunc expects data of type int, got %T", data)
	}

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	userToJson, _ := json.MarshalIndent(&user, "", "  ")

	log.Printf("User Information: %s", string(userToJson))
	return &user, nil
}
