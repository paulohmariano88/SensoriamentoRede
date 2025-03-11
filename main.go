package main

import (
	"fmt"
	"log"
	"net/http"
	"networksensor/controller"
	"networksensor/middleware"
)

func main() {


	http.HandleFunc("/api/login", controller.LoginHandler) //EndPoint para obter JWT

	http.HandleFunc("/api/getpackets",      middleware.JWTMiddleware(controller.HandleGetPackets)) //Busca os pacotes no banco
	http.HandleFunc("/api/startscan",       middleware.JWTMiddleware(controller.StartScan))         //Inicia o scan e armazena no banco
	http.HandleFunc("/api/stopscan",        middleware.JWTMiddleware(controller.CancelMeasure))    //Parar o scan
	http.HandleFunc("/api/listinterface",   middleware.JWTMiddleware(controller.ListAllInterfaces)) // Listar as interfaces de rede

	port := ":8080"
	fmt.Println("servidor rodando em http:IP:" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
