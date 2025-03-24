package model

import (
	"context"
	"fmt"
	"networksensor/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	Id       int    `bson:"id"`
	Login    string `bson:"login"`
	Password string `bson:"password"`
}

func SaveUser(user User) {
	// Conexão com o MongoDB
	client, db, err := database.ConectMongoDB()
	if err != nil {
		fmt.Println("Erro ao conectar no banco:", err)
		return
	}
	defer client.Disconnect(context.TODO()) // fecha a conexão ao fim

	// Collection
	collection := db.Collection("users")

	// Contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Documento que será inserido
	doc := bson.M{
		"id":       user.Id,
		"login":    user.Login,
		"password": user.Password,
	}

	// Inserção
	_, err = collection.InsertOne(ctx, doc)
	if err != nil {
		fmt.Println("Erro ao inserir usuário:", err)
		return
	}

	fmt.Println("Usuário salvo com sucesso!")
}