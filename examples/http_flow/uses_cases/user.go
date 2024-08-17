package use_cases

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Address  Address `json:"address"`
}

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
func GetUserFunc(userID int) (*User, error) {
	if userID == 0 {
		return nil, errors.New("error to retrieve user")
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

	fmt.Printf("User Information: %s\n", string(userToJson))
	return &user, nil
}
