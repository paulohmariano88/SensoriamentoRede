package main

import (
	"fmt"
	"net/http"
	"networksensor/controller"
	"networksensor/middleware"
)

func main() {

	http.HandleFunc("/api/login", controller.LoginHandler) //EndPoint para obter JWT

	http.HandleFunc("/api/getpackets", controller.HandleGetPackets)                                 //Busca os pacotes no banco
	http.HandleFunc("/api/startscan", middleware.JWTMiddleware(controller.StartScan))               //Inicia o scan e armazena no banco
	http.HandleFunc("/api/stopscan", middleware.JWTMiddleware(controller.CancelMeasure))            //Parar o scan
	http.HandleFunc("/api/listinterface", middleware.JWTMiddleware(controller.ListAllInterfaces))   // Listar as interfaces de rede
	http.HandleFunc("/api/getpacketsbydate", middleware.JWTMiddleware(controller.GetMeasureByDate)) //Buscar os dados pela data

	port := ":8383"
	fmt.Println("servidor rodando em http:IP:" + port)
	error := http.ListenAndServe(port, nil)
	if error != nil {
		fmt.Println("Erro encontrado em: ", error)
	} else {
		fmt.Println("Rodando com sucesso")
	}
}
