package model

import (
	"context"
	"fmt"
	"log"
	"networksensor/database"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Estrutura para armazenar pacotes enviados e calcular a latência
var packetTimes sync.Map
var cancelCapture context.CancelFunc //Cancela a captura de dados

type Measure struct {
	TimeStamp     time.Time `bson:"timestamp" json:"timestamp"`
	SourceIP      string    `bson:"source_ip" json:"source_ip"`
	DestinationIP string    `bson:"destination_ip" json:"destination_ip"`
	Protocol      string    `bson:"protocol" json:"protocol"`
	PacketSize    int       `bson:"packet_size" json:"packet_size"`
	LatencyUs     float64   `bson:"latency_us" json:"latency_us"`
}

func StartMeasure(interfaceName string) {
	// Interface de rede
	//interfaceName := `\Device\NPF_{11331855-82C3-4F81-B013-FDAD5B2D1DE2}`
	client, db, err := database.ConectMongoDB()
	//Cria um contexto de execução
	ctx, cancel := context.WithCancel(context.Background())
	cancelCapture = cancel //Armazena funçao de cancelamento

	if err != nil {
		fmt.Println("Erro ao conectar no Banco!!!")
	}

	defer database.DisconnectMongoDB(client)

	//cria o canal com um buffer de 100
	ch := make(chan Measure, 100)

	//Processa e salva os pacotes.
	go savePacketToMongoDB(db, ch)

	// Abrir a interface para captura de pacotes
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal("❌ Erro ao abrir a interface:", err)
	}
	defer handle.Close()

	// Criar o analisador de pacotes
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	fmt.Println("🔍 Escutando pacotes na interface:", interfaceName)

	// Loop para capturar pacotes em tempo real
	for {
		select {
		case <-ctx.Done():
			close(ch)
			return
		case packet := <-packetSource.Packets():
			go processPacket(packet, ch)

		}
	}
}

// Função que processa cada pacote capturado
func processPacket(packet gopacket.Packet, ch chan Measure) {
	// Obter timestamp e tamanho do pacote
	timestamp := packet.Metadata().Timestamp
	packetSize := packet.Metadata().Length

	// Identificação do protocolo (IPv4, IPv6, TCP, UDP)
	networkLayer := packet.NetworkLayer()
	transportLayer := packet.TransportLayer()

	if networkLayer == nil || transportLayer == nil {
		return
	}

	// Capturar IP de origem e destino
	sourceIP, destinationIP := networkLayer.NetworkFlow().Endpoints()
	protocol := transportLayer.LayerType().String()

	// Criar a estrutura com os dados do pacote
	measure := Measure{
		TimeStamp:     timestamp,
		SourceIP:      sourceIP.String(),
		DestinationIP: destinationIP.String(),
		Protocol:      protocol,
		PacketSize:    packetSize,
		LatencyUs:     0,
	}

	// Armazena o tempo do pacote enviado
	packetTimes.Store(measure.SourceIP, measure.TimeStamp)

	// Calcula latência se o pacote de resposta for recebido
	if value, ok := packetTimes.Load(measure.DestinationIP); ok {
		sentTime := value.(time.Time)
		latency := measure.TimeStamp.Sub(sentTime)

		// Convertendo para microsegundos (us)
		measure.LatencyUs = float64(latency.Seconds() * 1e6)
	}

	// 🔥 Exibe os detalhes do pacote capturado 🔥
	fmt.Println("📦 PACOTE CAPTURADO")
	fmt.Println("⏱ Tempo:", measure.TimeStamp)
	fmt.Println("📍 IP Origem:", measure.SourceIP)
	fmt.Println("🎯 IP Destino:", measure.DestinationIP)
	fmt.Println("📡 Protocolo:", measure.Protocol)
	fmt.Println("📏 Tamanho do Pacote:", measure.PacketSize, "bytes")
	fmt.Println("⚡ Latência:", measure.LatencyUs, "us")
	fmt.Println("--------------------------------------------------")

	ch <- measure

}

// Armazena no banco de dados
func savePacketToMongoDB(db *mongo.Database, ch chan Measure) {

	collection := database.GetCollection(db, "latency_data")

	for packet := range ch {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, packet)
		if err != nil {
			log.Printf("Erro ao salvar no MongDB: %v", err)
		} else {
			fmt.Println("Salvo no Mongo:", packet.SourceIP, "->", packet.DestinationIP)
		}
	}
}

// Busca os dados no banco de dados.
func FindAllPackets() ([]Measure, error) {

	_, db, err := database.ConectMongoDB()

	if err != nil {
		fmt.Println("Erro ao conectar ao banco de dados")
		return nil, err
	}

	collection := database.GetCollection(db, "latency_data")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Criando cursor para percorrer os dados
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal("Erro ao buscar dados: ", err)
		return nil, fmt.Errorf("erro ao buscar os dados: %v", err)
	}
	defer cursor.Close(ctx)

	var results []Measure

	//Iterar sobre os resultados e armazena-los
	for cursor.Next(ctx) {
		var packet Measure

		if err := cursor.Decode(&packet); err != nil {
			log.Println("Erro ao decodificar o banco:", err)
			continue
		}
		results = append(results, packet)
	}

	return results, nil
}

func StopMeasure() {
	if cancelCapture != nil {
		cancelCapture() //Cancela o Contexto
		fmt.Println()
	} else {
		fmt.Print(" Nenhma captura em execução!!")
	}
}
