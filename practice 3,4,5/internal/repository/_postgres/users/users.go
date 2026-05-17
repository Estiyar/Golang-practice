package users

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"practice3go/internal/repository/_postgres"
	"practice3go/pkg/modules"
)

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: 5 * time.Second,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var users []modules.User
	err := r.db.DB.SelectContext(ctx, &users, `
		SELECT id, name, email, age, gender, birth_date, created_at
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var u modules.User
	err := r.db.DB.GetContext(ctx, &u, `
		SELECT id, name, email, age, gender, birth_date, created_at
		FROM users
		WHERE id = $1
	`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, modules.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *Repository) CreateUser(u modules.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var id int
	err := r.db.DB.QueryRowxContext(ctx, `
		INSERT INTO users (name, email, age, gender, birth_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, u.Name, u.Email, u.Age, u.Gender, u.BirthDate).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateUser(id int, u modules.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	res, err := r.db.DB.ExecContext(ctx, `
		UPDATE users
		SET name = $1, email = $2, age = $3, gender = $4, birth_date = $5
		WHERE id = $6
	`, u.Name, u.Email, u.Age, u.Gender, u.BirthDate, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return modules.ErrUserNotFound
	}

	return nil
}

func (r *Repository) DeleteUserByID(id int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	res, err := r.db.DB.ExecContext(ctx, `
		DELETE FROM users WHERE id = $1
	`, id)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if affected == 0 {
		return 0, modules.ErrUserNotFound
	}

	return affected, nil
}

func (r *Repository) GetPaginatedUsers(page int, pageSize int, filters map[string]string, orderBy string) (modules.PaginatedResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	allowedOrder := map[string]string{
		"id_asc":          "id ASC",
		"id_desc":         "id DESC",
		"name_asc":        "name ASC",
		"name_desc":       "name DESC",
		"email_asc":       "email ASC",
		"email_desc":      "email DESC",
		"birth_date_asc":  "birth_date ASC",
		"birth_date_desc": "birth_date DESC",
	}

	orderClause, ok := allowedOrder[orderBy]
	if !ok {
		orderClause = "id ASC"
	}

	baseWhere := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if v := strings.TrimSpace(filters["id"]); v != "" {
		baseWhere = append(baseWhere, fmt.Sprintf("id = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(filters["name"]); v != "" {
		baseWhere = append(baseWhere, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(filters["email"]); v != "" {
		baseWhere = append(baseWhere, fmt.Sprintf("email ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(filters["gender"]); v != "" {
		baseWhere = append(baseWhere, fmt.Sprintf("gender ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(filters["birth_date"]); v != "" {
		baseWhere = append(baseWhere, fmt.Sprintf("DATE(birth_date) = $%d", argPos))
		args = append(args, v)
		argPos++
	}

	whereClause := strings.Join(baseWhere, " AND ")

	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	if err := r.db.DB.QueryRowxContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return modules.PaginatedResponse{}, err
	}

	queryArgs := append(args, pageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, name, email, age, gender, birth_date, created_at
		FROM users
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argPos, argPos+1)

	var users []modules.User
	if err := r.db.DB.SelectContext(ctx, &users, query, queryArgs...); err != nil {
		return modules.PaginatedResponse{}, err
	}

	return modules.PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCommonFriends(userID int, otherUserID int) ([]modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	if userID == otherUserID {
		return []modules.User{}, nil
	}

	query := `
		SELECT u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at
		FROM user_friends uf1
		JOIN user_friends uf2 ON uf1.friend_id = uf2.friend_id
		JOIN users u ON u.id = uf1.friend_id
		WHERE uf1.user_id = $1
		  AND uf2.user_id = $2
		ORDER BY u.id
	`

	var users []modules.User
	if err := r.db.DB.SelectContext(ctx, &users, query, userID, otherUserID); err != nil {
		return nil, err
	}

	return users, nil
}

func parseIntSafe(value string, def int) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return n
}