package models

type User struct {
	ID         int64  `json:"id" example:"1"`
	PublicId   int64  `json:"public_id" example:"87610"`
	FirstName  string `json:"first_name" example:"John"`
	LastName   string `json:"last_name" example:"Doe"`
	FatherName string `json:"father_name" example:"Smith"`
	Email      string `json:"email" example:"n2OjP@example.com"`
	Password   string `json:"-"`
}
