package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	gatewayApiURL = "http://%s:82/smartmeter/api/read"
)

type ApiResponse struct {
	FirmwareRunning         bool
	FirmwareAvailable       bool
	FirmwareUpdateAvailable bool
	ElectricityTariff       int
	EnergyDeliveredTariff1  float64
	EnergyReturnedTariff1   float64
	EnergyDeliveredTariff2  float64
	EnergyReturnedTariff2   float64
	PowerDeliveredTotal     int
	PowerReturnedTotal      int
	PowerDeliveredL1        int
	PowerDeliveredL2        int
	PowerDeliveredL3        int
	PowerReturnedL1         int
	PowerReturnedL2         int
	PowerReturnedL3         int
	VoltageL1               int
	VoltageL2               int
	VoltageL3               int
	CurrentL1               int
	CurrentL2               int
	CurrentL3               int
	GasDelivered            float64
	GasDeliveredHour        float64
	PowerDelivered          float64
	PowerDeliveredHour      float64
}

type Exporter struct {
}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	//ch <- descSystemConnected
}

func main() {
	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	log.Println("⚙️ Exporter is ready to accept requests")

	fetchSystemData()

	//log.Fatal(http.ListenAndServe(":9000", nil))
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	fetchSystemData()
}

func fetchSystemData() {
	host := os.Getenv("SGPE_HOST")
	if host == "" {
		log.Fatal("💥 SGPE_HOST not set")
	}

	getDataFromApi(fmt.Sprintf(gatewayApiURL, host))

	//var envoyData = envoyType{
	//	ProductionType:          &envoyProductionType{},
	//	ProductionInvertersType: &[]envoyProductionInvertersType{},
	//	HomeType:                &envoyHomeType{},
	//}
	//
	//return envoyData
}

func getDataFromApi(url string, data interface{}) {
	req, _ := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(nil))

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("⚠️ Request %s failed\n", url)
		log.Fatal(err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = fmt.Errorf("️⚠️ Request failed with status %d\n", response.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, _ := io.ReadAll(response.Body)

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal("💥 JSON object was not valid")
	}
}
