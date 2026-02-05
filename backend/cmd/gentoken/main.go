package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         "22222222-2222-2222-2222-222222222222",
		"email":           "admin@goplan.dev",
		"role":            "admin",
		"organization_id": "11111111-1111-1111-1111-111111111111",
	})

	tokenString, err := token.SignedString([]byte("+foDnwv0f8Ya3BX5GIsO/N3I023XuLJjVV5m1KDQMoc="))
	if err != nil {
		panic(err)
	}
	fmt.Println(tokenString)
}
