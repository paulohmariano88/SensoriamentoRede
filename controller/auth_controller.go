package controller

import (
	"encoding/json"
	"net/http"
	"networksensor/middleware"
)

//Estrutura de capturar credenciais
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}


// Banco de dados
var users = map[string]string {
	"admin": "1234",
	"user": "senha",
}


// Handler de autenticação
func LoginHandler(w http.ResponseWriter, r *http.Request){

	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	//Verifica se o usuário existe e a senha está correta
	if password, ok := users[creds.Username]; !ok || password != creds.Password {
		http.Error(w, "Usuário ou senha inválidas", http.StatusUnauthorized)
		return
	}

	//Gerar um token JWT
	token, err := middleware.GenerateJWT(creds.Username)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	//Retornar o token como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}


