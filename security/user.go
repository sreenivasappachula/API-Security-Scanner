package security

// JSON type
type JSON map[string]interface{}

// User struct
type User struct {
	Email    string
	Password string
}

// Setter methods

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetPassword(password string) {
	u.Password = password
}

// Return email + password in JSON format
func (u *User) GetJSON() JSON {

	return JSON{
		"email":    u.Email,
		"password": u.Password,
	}
}
