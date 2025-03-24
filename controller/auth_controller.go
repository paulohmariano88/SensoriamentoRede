package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"networksensor/database"
	"networksensor/middleware"
	"networksensor/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Conectar ao MongoDB
	client, db, err := database.ConectMongoDB()
	if err != nil {
		http.Error(w, "Erro ao conectar no banco de dados", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Buscar o usuário pela coleção
	collection := db.Collection("users")
	var user struct {
		Login    string `bson:"login"`
		Password string `bson:"password"`
	}

	err = collection.FindOne(ctx, bson.M{"login": creds.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "Usuário ou senha inválidas", http.StatusUnauthorized)
		return
	}

	// Verificar senha
	if user.Password != creds.Password {
		http.Error(w, "Usuário ou senha inválidas", http.StatusUnauthorized)
		return
	}

	// Gerar token JWT
	token, err := middleware.GenerateJWT(creds.Username)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	// Retornar o token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func SaveUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var newUser model.User

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Validação básica
	if newUser.Login == "" || newUser.Password == "" {
		http.Error(w, "Login e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	// Conexão com MongoDB
	client, db, err := database.ConectMongoDB()
	if err != nil {
		http.Error(w, "Erro ao conectar no banco", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("users")

	// Verificar se o usuário já existe
	var existingUser bson.M
	err = collection.FindOne(ctx, bson.M{"login": newUser.Login}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Usuário já existe", http.StatusConflict)
		return
	}

	// Inserir novo usuário
	_, err = collection.InsertOne(ctx, bson.M{
		"login":    newUser.Login,
		"password": newUser.Password, // ⚠️ em produção: use hash!
	})

	if err != nil {
		http.Error(w, "Erro ao salvar usuário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Usuário criado com sucesso!"})
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	client, db, err := database.ConectMongoDB()
	if err != nil {
		http.Error(w, "Erro ao conectar no banco", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("users")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Erro ao buscar usuários", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		err := cursor.Decode(&user)
		if err != nil {
			fmt.Println("Erro ao decodificar usuário:", err)
			continue
		}
		// Evita retornar senha (opcional, mas recomendado)
		user.Password = ""
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
