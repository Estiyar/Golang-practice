package modules

import "time"

type User struct {
	ID        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     *string    `db:"email" json:"email"`
	Age       *int       `db:"age" json:"age"`
	Gender    *string    `db:"gender" json:"gender"`
	BirthDate *time.Time `db:"birth_date" json:"birth_date"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type PaginatedResponse struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}