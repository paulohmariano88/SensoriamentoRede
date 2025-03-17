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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Estrutura para armazenar pacotes enviados e calcular a latência
var packetTimes sync.Map
var cancelCapture context.CancelFunc // Cancela a captura de dados

type Measure struct {
	TimeStamp     time.Time `bson:"timestamp" json:"timestamp"`
	SourceIP      string    `bson:"source_ip" json:"source_ip"`
	DestinationIP string    `bson:"destination_ip" json:"destination_ip"`
	Protocol      string    `bson:"protocol" json:"protocol"`
	PacketSize    int       `bson:"packet_size" json:"packet_size"`
	LatencyUs     float64   `bson:"latency_us" json:"latency_us"`
}

// Inicia a medição
func StartMeasure(interfaceName string) {
	// Interface de rede
	client, db, err := database.ConectMongoDB()
	if err != nil {
		fmt.Println("Erro ao conectar no banco:", err)
		return
	}

	// Cria um contexto de execução
	ctx, cancel := context.WithCancel(context.Background())
	cancelCapture = cancel // Armazena função de cancelamento
	defer database.DisconnectMongoDB(client)

	// Canal para comunicação entre goroutines
	ch := make(chan Measure, 100)
	go savePacketToMongoDB(db, ch)

	// Abrir a interface para captura de pacotes
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		fmt.Println("Erro ao abrir a interface:", err)
		return
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
			fmt.Println("⏹️  Captura interrompida.")
			return
		case packet := <-packetSource.Packets():
			go processPacket(packet, ch)
		}
	}
}

// Processa pacotes capturados
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

	// Armazena apenas pacotes enviados (evita sobrescrever timestamp prematuramente)
	_, exists := packetTimes.LoadOrStore(measure.SourceIP, measure.TimeStamp)
	if !exists {
		packetTimes.Store(measure.SourceIP, measure.TimeStamp)
	}

	// Calcula latência apenas se o pacote de resposta foi registrado
	if value, ok := packetTimes.Load(measure.DestinationIP); ok {
		sentTime := value.(time.Time)
		if measure.TimeStamp.After(sentTime) {
			latency := measure.TimeStamp.Sub(sentTime)
			measure.LatencyUs = float64(latency.Microseconds())
		} else {
			// Se o pacote chegou antes do envio, ignoramos
			fmt.Println("⚠️ Pacote recebido antes do envio, ignorando latência")
			measure.LatencyUs = 0
		}
	}

	// 🔥 Exibe os detalhes do pacote capturado 🔥
	fmt.Printf("📡 [%s] %s -> %s | 🕒 %d µs\n", measure.Protocol, measure.SourceIP, measure.DestinationIP, measure.LatencyUs)
	ch <- measure
}

// Salva pacotes no banco de dados
func savePacketToMongoDB(db *mongo.Database, ch chan Measure) {
	collection := database.GetCollection(db, "latency_data")

	for packet := range ch {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, packet)
		if err != nil {
			log.Printf("Erro ao salvar no MongoDB: %v", err)
		} else {
			fmt.Println("Salvo no MongoDB:", packet.SourceIP, "->", packet.DestinationIP)
		}
	}
}

// Busca pacotes no banco de dados
func FindAllPackets() ([]Measure, error) {
	_, db, err := database.ConectMongoDB()
	if err != nil {
		fmt.Println("Erro ao conectar ao banco de dados!")
		return nil, err
	}
	defer database.DisconnectMongoDB(db.Client())

	collection := database.GetCollection(db, "latency_data")

	findOptions := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(1000) //Limita o registro a 1000

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Erro ao buscar dados: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Measure
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Erro ao decodificar documentos: %v", err)
		return nil, err
	}

	return results, nil
}

// Busca pacotes dentro de um intervalo de tempo
func GetMeasureByDate(begin time.Time, end time.Time) ([]Measure, error) {
	client, db, err := database.ConectMongoDB()
	if err != nil {
		fmt.Println("Erro ao conectar ao banco de dados!")
		return nil, err
	}
	defer database.DisconnectMongoDB(client)

	collection := database.GetCollection(db, "latency_data")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(1000) //Limita o registro a 1000

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": begin, // Maior ou igual a `begin`
			"$lte": end,   // Menor ou igual a `end`
		},
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Erro ao buscar dados por período: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Measure
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Erro ao decodificar documentos: %v", err)
		return nil, err
	}

	return results, nil
}

// Para a captura de pacotes
func StopMeasure() {
	if cancelCapture != nil {
		cancelCapture()
		fmt.Println("Captura interrompida.")
	} else {
		fmt.Println(" Nenhuma captura em execução.")
	}
}
