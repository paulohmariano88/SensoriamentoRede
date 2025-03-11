package main

import (
	"fmt"
	"log"
	"net/http"
	"networksensor/controller"
)

func main() {

	http.HandleFunc("/api/getpackets", controller.HandleGetPackets) //Busca os pacotes no banco
	http.HandleFunc("/api/startscan", controller.StartScan)         //Inicia o scan e armazena no banco
	http.HandleFunc("/api/stopscan", controller.CancelMeasure)    //Parar o scan
	http.HandleFunc("/api/listinterface", controller.ListAllInterfaces)

	port := ":8080"
	fmt.Println("servidor rodando em http:IP:" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
