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
	gatewayApiURL = "http://%s/smartmeter/api/read"
)

type ApiResponse struct {
	FirmwareRunning         string `json:"firmware_running"`
	FirmwareAvailable       string `json:"firmware_available"`
	FirmwareUpdateAvailable string `json:"firmware_update_available"`
	ElectricityTariff       string `json:"ElectricityTariff"`
	EnergyDeliveredTariff1  string `json:"EnergyDeliveredTariff1"`
	EnergyReturnedTariff1   string `json:"EnergyReturnedTariff1"`
	EnergyDeliveredTariff2  string `json:"EnergyDeliveredTariff2"`
	EnergyReturnedTariff2   string `json:"EnergyReturnedTariff2"`
	PowerDeliveredTotal     string `json:"PowerDelivered_total"`
	PowerReturnedTotal      string `json:"PowerReturned_total"`
	PowerDeliveredL1        string `json:"PowerDelivered_l1"`
	PowerDeliveredL2        string `json:"PowerDelivered_l2"`
	PowerDeliveredL3        string `json:"PowerDelivered_l3"`
	PowerReturnedL1         string `json:"PowerReturned_l1"`
	PowerReturnedL2         string `json:"PowerReturned_l2"`
	PowerReturnedL3         string `json:"PowerReturned_l3"`
	VoltageL1               string `json:"Voltage_l1"`
	VoltageL2               string `json:"Voltage_l2"`
	VoltageL3               string `json:"Voltage_l3"`
	CurrentL1               string `json:"Current_l1"`
	CurrentL2               string `json:"Current_l2"`
	CurrentL3               string `json:"Current_l3"`
	PowerDeliveredHour      string `json:"PowerDeliveredHour"`
	PowerDeliveredNet       string `json:"PowerDeliveredNetto"`
	GasDelivered            string `json:"GasDelivered"`
	GasDeliveredHour        string `json:"GasDeliveredHour"`
}

type Stats struct {
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

	var apiResponse ApiResponse
	//var stats *Stats
	getDataFromApi(fmt.Sprintf(gatewayApiURL, host), &apiResponse)

	fmt.Println(apiResponse)
}

func getDataFromApi(url string, data *ApiResponse) {
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

	err = json.Unmarshal(body, data)
	if err != nil {
		log.Fatalf("💥 JSON object was not valid: %s", err)
	}
}
