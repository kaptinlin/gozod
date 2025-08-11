package testdata

import "time"

// User represents a simple user structure for testing
type User struct {
	ID        string    `json:"id" gozod:"required,uuid"`
	Name      string    `json:"name" gozod:"required,min=2,max=50"`
	Email     string    `json:"email" gozod:"required,email"`
	Age       int       `json:"age" gozod:"required,min=18,max=120"`
	Status    string    `json:"status" gozod:"enum=active inactive,default=active"`
	CreatedAt time.Time `json:"created_at" gozod:"required"`
}
