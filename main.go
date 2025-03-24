package main

import (
	"fmt"
	"net/http"
	"networksensor/controller"
	"networksensor/middleware"
)

func main() {

	// Envolvendo todas as rotas com o middleware de CORS
	http.Handle("/api/login", middleware.CORSMiddleware(http.HandlerFunc(controller.LoginHandler)))
	http.Handle("/api/saveUser", middleware.CORSMiddleware(http.HandlerFunc(controller.SaveUserHandler)))
	http.Handle("/api/getUsers", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.GetUsersHandler))))
	http.Handle("/api/getpackets", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.HandleGetPackets))))
	http.Handle("/api/startscan", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.StartScan))))
	http.Handle("/api/stopscan", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.CancelMeasure))))
	http.Handle("/api/listinterface", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.ListAllInterfaces))))
	http.Handle("/api/getpacketsbydate", middleware.CORSMiddleware(middleware.JWTMiddleware(http.HandlerFunc(controller.GetMeasureByDate))))

	port := ":8383"
	fmt.Println("servidor rodando em http:IP:" + port)
	error := http.ListenAndServe(port, nil)
	if error != nil {
		fmt.Println("Erro encontrado em: ", error)
	} else {
		fmt.Println("Rodando com sucesso")
	}
}
