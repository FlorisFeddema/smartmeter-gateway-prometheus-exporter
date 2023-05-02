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
	"strconv"
)

var (
	gatewayApiURL = "http://%s/smartmeter/api/read"

	descFirmwareRunning         = prometheus.NewDesc("smartmeter_gateway_firmware_running", "Current version of the running firmware", nil, nil)
	descFirmwareAvailable       = prometheus.NewDesc("smartmeter_gateway_firmware_available", "Latest avaialble version fo the firmware", nil, nil)
	descFirmwareUpdateAvailable = prometheus.NewDesc("smartmeter_gateway_firmware_update_available", "If there is a new version of the firmware available", nil, nil)
	descGasConsumed             = prometheus.NewDesc("smartmeter_gateway_gas_consumed", "The total amount of gas that is consumed cubic meters", nil, nil)
	descGasConsumedHour         = prometheus.NewDesc("smartmeter_gateway_gas_consumed_hour", "The amount of gas consumed in the current hour in cubic meters", nil, nil)
	//desc                        = prometheus.NewDesc("smartmeter_gateway_", "", nil, nil)
)

type Firmware struct {
	Running         int
	Available       int
	UpdateAvailable bool
}
type Gas struct {
	Consumed     float64
	ConsumedHour float64
}

type Power struct {
	Tariff          int
	ConsumedTariff1 float64
	ProducedTariff1 float64
	ConsumedTariff2 float64
	ProducedTariff2 float64
	ConsumedTotal   int
	ProducedTotal   int
	ConsumedL1      int
	ConsumedL2      int
	ConsumedL3      int
	ProducedL1      int
	ProducedL2      int
	ProducedL3      int
	VoltageL1       int
	VoltageL2       int
	VoltageL3       int
	CurrentL1       int
	CurrentL2       int
	CurrentL3       int
	ConsumedHour    float64
}
type Stats struct {
	Firmware Firmware
	Gas      Gas
	Power    Power
}

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

func getInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func getBool(s string) bool {
	value, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func getFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func fetchSystemData() {
	host := os.Getenv("SGPE_HOST")
	if host == "" {
		log.Fatal("💥 SGPE_HOST not set")
	}

	var apiResponse *ApiResponse
	getDataFromApi(fmt.Sprintf(gatewayApiURL, host), &apiResponse)

	stats := Stats{
		Firmware: Firmware{
			Running:         getInt(apiResponse.FirmwareRunning),
			Available:       getInt(apiResponse.FirmwareAvailable),
			UpdateAvailable: getBool(apiResponse.FirmwareUpdateAvailable),
		},
		Gas: Gas{
			Consumed:     getFloat(apiResponse.GasDelivered),
			ConsumedHour: getFloat(apiResponse.GasDeliveredHour),
		},
		Power: Power{
			Tariff:          getInt(apiResponse.ElectricityTariff),
			ConsumedTariff1: getFloat(apiResponse.EnergyDeliveredTariff1),
			ProducedTariff1: getFloat(apiResponse.EnergyReturnedTariff1),
			ConsumedTariff2: getFloat(apiResponse.EnergyDeliveredTariff2),
			ProducedTariff2: getFloat(apiResponse.EnergyReturnedTariff2),
			ConsumedTotal:   getInt(apiResponse.PowerDeliveredTotal),
			ProducedTotal:   getInt(apiResponse.PowerReturnedTotal),
			ConsumedL1:      getInt(apiResponse.PowerDeliveredL1),
			ConsumedL2:      getInt(apiResponse.PowerDeliveredL2),
			ConsumedL3:      getInt(apiResponse.PowerDeliveredL3),
			ProducedL1:      getInt(apiResponse.PowerReturnedL1),
			ProducedL2:      getInt(apiResponse.PowerReturnedL2),
			ProducedL3:      getInt(apiResponse.PowerReturnedL3),
			VoltageL1:       getInt(apiResponse.VoltageL1),
			VoltageL2:       getInt(apiResponse.VoltageL2),
			VoltageL3:       getInt(apiResponse.VoltageL3),
			CurrentL1:       getInt(apiResponse.CurrentL1),
			CurrentL2:       getInt(apiResponse.CurrentL2),
			CurrentL3:       getInt(apiResponse.CurrentL3),
			ConsumedHour:    getFloat(apiResponse.PowerDeliveredHour),
		},
	}

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

	err = json.Unmarshal(body, data)
	if err != nil {
		log.Fatalf("💥 JSON object was not valid: %s", err)
	}
}
