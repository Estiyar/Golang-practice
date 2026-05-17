package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"practice3go/internal/usecase"
	"practice3go/pkg/modules"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (h *UserHandler) Users(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/users" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		orderBy := r.URL.Query().Get("order_by")

		filters := map[string]string{
			"id":         r.URL.Query().Get("id"),
			"name":       r.URL.Query().Get("name"),
			"email":      r.URL.Query().Get("email"),
			"gender":     r.URL.Query().Get("gender"),
			"birth_date": r.URL.Query().Get("birth_date"),
		}

		resp, err := h.uc.GetPaginatedUsers(page, pageSize, filters, orderBy)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPost:
		var req struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			Age       int    `json:"age"`
			Gender    string `json:"gender"`
			BirthDate string `json:"birth_date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}

		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "birth_date must be YYYY-MM-DD")
			return
		}

		id, err := h.uc.CreateUser(modules.User{
		Name: req.Name,
		Email: &req.Email,
		Age: &req.Age,
		Gender: &req.Gender,
		BirthDate: &birthDate,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]int{"id": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *UserHandler) UserByID(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/users/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := h.uc.GetUserByID(id)
		if err != nil {
			if errors.Is(err, modules.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, user)

	case http.MethodPut:
		var req struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			Age       int    `json:"age"`
			Gender    string `json:"gender"`
			BirthDate string `json:"birth_date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}

		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "birth_date must be YYYY-MM-DD")
			return
		}

		err = h.uc.UpdateUser(id, modules.User{
		Name:      req.Name,
		Email:     &req.Email,
		Age:       &req.Age,
		Gender:    &req.Gender,
		BirthDate: &birthDate,
		})
		if err != nil {
			if errors.Is(err, modules.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		_, err := h.uc.DeleteUser(id)
		if err != nil {
			if errors.Is(err, modules.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *UserHandler) CommonFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
	otherUserID, _ := strconv.Atoi(r.URL.Query().Get("other_user_id"))

	if userID <= 0 || otherUserID <= 0 {
		writeError(w, http.StatusBadRequest, "user_id and other_user_id are required")
		return
	}

	users, err := h.uc.GetCommonFriends(userID, otherUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}