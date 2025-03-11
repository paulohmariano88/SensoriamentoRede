package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoURI = "mongodb://localhost:27017"
const databaseName = "db_networkmonitor"

func ConectMongoDB() (*mongo.Client, *mongo.Database, error) {

	var err error

	//Criar contexto com timeout para evitar conexões travadas
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Configurar o cliente do MongoDb
	clientOption := options.Client().ApplyURI(mongoURI)

	//conectar ao mongodb
	client, err := mongo.Connect(ctx, clientOption)
	if err != nil {
		log.Printf("Erro ao conectar ao MOngoDB: %v", err)
		return nil, nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Erro ao testar conexaõ com MongoDB: %v", err)
		return nil, nil, err
	}

	// Criar instância do Banco de dados
	db := client.Database(databaseName)
	fmt.Print("Conectado ao Mongo DB")

	return client, db, nil
}

func DisconnectMongoDB(client *mongo.Client) {

	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Disconnect(ctx)
		if err != nil {
			log.Printf(" Erro ao desconectar do Mongo: %v", err)
		} else {
			fmt.Println("Conexão com MongoDB encerrada com sucesso!!")
		}
	}
}

func GetCollection(db *mongo.Database, collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}
