package main

// Custom JSON type
type JSON map[string]interface{}

// User struct
type User struct {
	email    string
	password string
}

// Setter methods

func (u *User) SetEmail(email string) {
	u.email = email
}

func (u *User) SetPassword(password string) {
	u.password = password
}

// Return email + password in JSON format
func (u *User) GetJSON() JSON {

	return JSON{
		"email":    u.email,
		"password": u.password,
	}
}
