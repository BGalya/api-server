package api_sec

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	UserID   string `json:"user_id"` // unique id created in Register func
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func Register(w http.ResponseWriter, r *http.Request) {
	rec := &responseRecorder{ResponseWriter: w}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		WriteToFile(r, http.StatusMethodNotAllowed, rec.bodyLen)
		return
	}
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		WriteToFile(r, http.StatusBadRequest, rec.bodyLen)
		return
	}
	user.ID = len(users) + 1
	users = append(users, user)
	json.NewEncoder(w).Encode(user)

	WriteToFile(r, http.StatusOK, len(user.Username))
}

func Login(w http.ResponseWriter, r *http.Request) {
	rec := &responseRecorder{ResponseWriter: w}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		WriteToFile(r, http.StatusMethodNotAllowed, rec.bodyLen)
		return
	}
	var creds User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		WriteToFile(r, http.StatusBadRequest, rec.bodyLen)
		return
	}

	// Authenticate user
	var authenticatedUser *User
	for _, user := range users {
		if user.Username == creds.Username && user.Password == creds.Password {
			authenticatedUser = &user
			break
		}
	}
	if authenticatedUser == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		WriteToFile(r, http.StatusUnauthorized, rec.bodyLen)
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		UserID:   strconv.Itoa(authenticatedUser.ID), // Store the unique user ID
		Username: authenticatedUser.Username,
		Role:     authenticatedUser.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		WriteToFile(r, http.StatusInternalServerError, rec.bodyLen)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	WriteToFile(r, http.StatusOK, len(tokenString))
}

func AccountsHandler(w http.ResponseWriter, r *http.Request, claims *Claims) {
	rec := &responseRecorder{ResponseWriter: w}
	if claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		WriteToFile(r, http.StatusForbidden, rec.bodyLen)
		return
	}
	if r.Method == http.MethodPost {
		createAccount(w, r, claims)
		WriteToFile(r, http.StatusOK, 0)
		return
	}
	if r.Method == http.MethodGet {
		listAccounts(w, r, claims)
		WriteToFile(r, http.StatusOK, 0)
		return
	}
}

func createAccount(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var acc Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	acc.ID = len(accounts) + 1
	acc.CreatedAt = time.Now()
	accounts = append(accounts, acc)
	json.NewEncoder(w).Encode(acc)
}

func listAccounts(w http.ResponseWriter, r *http.Request, claims *Claims) {
	json.NewEncoder(w).Encode(accounts)
}

func BalanceHandler(w http.ResponseWriter, r *http.Request, claims *Claims) {
	rec := &responseRecorder{ResponseWriter: w}
	if claims.Role != "user" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		WriteToFile(r, http.StatusForbidden, rec.bodyLen)
		return
	}
	switch r.Method {
	case http.MethodGet:
		getBalance(w, r, claims)
		WriteToFile(r, http.StatusOK, 0)
	case http.MethodPost:
		depositBalance(w, r, claims)
		WriteToFile(r, http.StatusOK, 0)
	case http.MethodDelete:
		withdrawBalance(w, r, claims)
		WriteToFile(r, http.StatusOK, 0)
	}
}

// Compare user_id from query with claims.UserID
// returns error if users do not match
func isAuthorizedUser(requestedUserID int, claims *Claims) bool {
	return strconv.Itoa(requestedUserID) == claims.UserID
}

func getBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	userId := r.URL.Query().Get("user_id")
	uid, err := strconv.Atoi(userId)
	// Invalid input
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isAuthorizedUser(uid, claims) {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	for _, acc := range accounts {
		if acc.UserID == uid {
			json.NewEncoder(w).Encode(map[string]float64{"balance": acc.Balance})
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func depositBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var body struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isAuthorizedUser(body.UserID, claims) {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	for i, acc := range accounts {
		if acc.UserID == body.UserID {
			accounts[i].Balance += body.Amount
			json.NewEncoder(w).Encode(accounts[i])
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func withdrawBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var body struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isAuthorizedUser(body.UserID, claims) {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	for i, acc := range accounts {
		if acc.UserID == body.UserID {
			if acc.Balance < body.Amount {
				http.Error(w, ErrInsufficientFunds.Error(), http.StatusBadRequest)
				return
			}
			accounts[i].Balance -= body.Amount
			json.NewEncoder(w).Encode(accounts[i])
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func Auth(next func(http.ResponseWriter, *http.Request, *Claims)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, claims)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bodyLen    int
}

// WriteHeader captures the status code
func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the body length
func (rec *responseRecorder) Write(data []byte) (int, error) {
	size, err := rec.ResponseWriter.Write(data)
	rec.bodyLen += size
	return size, err
}
