package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"networksensor/model"
	"github.com/google/gopacket/pcap"
)



func HandleGetPackets(w http.ResponseWriter, r *http.Request){

	packets, err := model.FindAllPackets()
	if err != nil {
		http.Error(w, "Erro ao buscar os pacotes", http.StatusInternalServerError)
		return
	}

	//Definição do cabeçalho para JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)


	//Codificar e Enviar os pacotes como JSON
	json.NewEncoder(w).Encode(packets)
}


func StartScan(w http.ResponseWriter, r *http.Request) {

  	var request model.Interface

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}
	
	fmt.Print("Iniciando sensor de rede")
	model.StartMeasure(request.NameInterface)
}


func ListAllInterfaces(w http.ResponseWriter, r *http.Request){

	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	var data []any 

	fmt.Println("Interfaces disponiveis")
	for _, device := range devices {
		fmt.Printf("- Nome:%s\n", device.Name)
		fmt.Printf(" Descrição: %s\n", device.Description)
	    it := model.Interface{
			NameInterface: device.Name,
			Description: device.Description,
			//IpRelational: device.Addresses,
		}		

		data = append(data, it)

		if len(device.Addresses) > 0 {
			fmt.Println(" Endereço IP associado:")
			for _, addr := range device.Addresses {
				fmt.Printf(" - %s\n", addr.IP)
			}
		}
	}

		//Definição do cabeçalho para JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	
		//Codificar e Enviar os pacotes como JSON
		json.NewEncoder(w).Encode(data)

}

func CancelMeasure(w http.ResponseWriter, r *http.Request){

	model.StopMeasure()

}