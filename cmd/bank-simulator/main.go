package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type authorizationRequest struct {
	CardNumber string `json:"card_number"`
}

type authorizationResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /payments", authorize)

	if err := http.ListenAndServe(":8080", mux); err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("bank simulator: %w", err))
	}
}

func authorize(w http.ResponseWriter, r *http.Request) {
	var request authorizationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.CardNumber == "" {
		http.Error(w, `{"error":"invalid authorization request"}`, http.StatusBadRequest)
		return
	}

	lastDigit := request.CardNumber[len(request.CardNumber)-1]
	if lastDigit == '0' {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	response := authorizationResponse{Authorized: (lastDigit-'0')%2 == 1}
	if response.Authorized {
		response.AuthorizationCode = "local-auth"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
