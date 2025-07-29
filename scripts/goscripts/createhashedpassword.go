package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "ToilAndTrouble123!" // <--- CHOOSE YOUR DESIRED PASSWORD HERE
	// $2a$10$ZKVVMvdtAFGawUIJy.p2Pe.N8ghhVjX2ZYre/nK0oNq7FvvfCZ/0i
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}
	fmt.Printf("Password: %s\nHashed Password: %s\n", password, string(hashedPassword))
}
