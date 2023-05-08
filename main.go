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
	descPowerTariff             = prometheus.NewDesc("smartmeter_gateway_power_tariff", "The current power tariff", nil, nil)
	descPowerConsumedTotal      = prometheus.NewDesc("smartmeter_gateway_power_consumed_total", "The total amount of power consumed in kWh", []string{"tariff"}, nil)
	descPowerProducedTotal      = prometheus.NewDesc("smartmeter_gateway_power_produced_total", "The total amount of power produced in kWh", []string{"tariff"}, nil)
	descPowerConsumedCurrent    = prometheus.NewDesc("smartmeter_gateway_power_consumed_current", "The current amount of power consumed in Watts", nil, nil)
	descPowerProducedCurrent    = prometheus.NewDesc("smartmeter_gateway_power_produced_current", "The current amount of power produced in Watts", nil, nil)
	descPowerConsumedPhase      = prometheus.NewDesc("smartmeter_gateway_power_consumed_phase", "The current amount of power consumed in Watts on a phase", []string{"phase"}, nil)
	descPowerProducedPhase      = prometheus.NewDesc("smartmeter_gateway_power_produced_phase", "The current amount of power produced in Watts on a phase", []string{"phase"}, nil)
	descPowerVoltagePhase       = prometheus.NewDesc("smartmeter_gateway_power_voltage_phase", "The current voltage on the phase", []string{"phase"}, nil)
	descPowerCurrentPhase       = prometheus.NewDesc("smartmeter_gateway_power_current_phase", "The current current on the phase", []string{"phase"}, nil)
	descPowerConsumedHour       = prometheus.NewDesc("smartmeter_gateway_power_consumed_hour", "The amount of power consumed last in this hour in kWh", nil, nil)
	descPowerConsumedNet        = prometheus.NewDesc("smartmeter_gateway_power_consumed_nett", "The net amount of power currently consumed in Watts", nil, nil)
)

type firmware struct {
	Running         int
	Available       int
	UpdateAvailable bool
}

type gas struct {
	Consumed     float64
	ConsumedHour float64
}

type power struct {
	Tariff               int
	ConsumedTotalTariff1 float64
	ProducedTotalTariff1 float64
	ConsumedTotalTariff2 float64
	ProducedTotalTariff2 float64
	ConsumedCurrent      int
	ProducedCurrent      int
	ConsumedPhase1       int
	ConsumedPhase2       int
	ConsumedPhase3       int
	ProducedPhase1       int
	ProducedPhase2       int
	ProducedPhase3       int
	VoltagePhase1        int
	VoltagePhase2        int
	VoltagePhase3        int
	CurrentPhase1        int
	CurrentPhase2        int
	CurrentPhase3        int
	ConsumedHour         float64
	ConsumedNet          float64
}

type Stats struct {
	Firmware firmware
	Gas      gas
	Power    power
}

type apiResponse struct {
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
	PowerDeliveredNetto     string `json:"PowerDeliveredNetto"`
	GasDelivered            string `json:"GasDelivered"`
	GasDeliveredHour        string `json:"GasDeliveredHour"`
}

type Exporter struct {
}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- descFirmwareRunning
	ch <- descFirmwareAvailable
	ch <- descFirmwareUpdateAvailable
	ch <- descGasConsumed
	ch <- descGasConsumedHour
	ch <- descPowerTariff
	ch <- descPowerConsumedTotal
	ch <- descPowerProducedTotal
	ch <- descPowerConsumedCurrent
	ch <- descPowerProducedCurrent
	ch <- descPowerConsumedPhase
	ch <- descPowerProducedPhase
	ch <- descPowerVoltagePhase
	ch <- descPowerCurrentPhase
	ch <- descPowerConsumedHour
	ch <- descPowerConsumedNet
}

func main() {
	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	log.Println("⚙️ Exporter is ready to accept requests")

	log.Fatal(http.ListenAndServe(":80", nil))
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	stats := fetchSystemData()

	metrics <- prometheus.MustNewConstMetric(
		descFirmwareRunning,
		prometheus.GaugeValue,
		float64(stats.Firmware.Running),
	)
	metrics <- prometheus.MustNewConstMetric(
		descFirmwareAvailable,
		prometheus.GaugeValue,
		float64(stats.Firmware.Available),
	)
	metrics <- prometheus.MustNewConstMetric(
		descFirmwareUpdateAvailable,
		prometheus.GaugeValue,
		boolToFloat64(stats.Firmware.UpdateAvailable),
	)
	metrics <- prometheus.MustNewConstMetric(
		descGasConsumed,
		prometheus.CounterValue,
		stats.Gas.Consumed,
	)
	metrics <- prometheus.MustNewConstMetric(
		descGasConsumedHour,
		prometheus.GaugeValue,
		stats.Gas.ConsumedHour,
	)

	metrics <- prometheus.MustNewConstMetric(
		descPowerTariff,
		prometheus.GaugeValue,
		float64(stats.Power.Tariff),
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedTotal,
		prometheus.CounterValue,
		stats.Power.ConsumedTotalTariff1,
		"1",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedTotal,
		prometheus.CounterValue,
		stats.Power.ConsumedTotalTariff2,
		"2",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedCurrent,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedCurrent),
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerProducedCurrent,
		prometheus.GaugeValue,
		float64(stats.Power.ProducedCurrent),
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase1),
		"1",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase2),
		"2",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase3),
		"3",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerProducedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ProducedPhase1),
		"1",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerProducedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ProducedPhase2),
		"2",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerProducedPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ProducedPhase3),
		"3",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerVoltagePhase,
		prometheus.GaugeValue,
		float64(stats.Power.VoltagePhase1),
		"1",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerVoltagePhase,
		prometheus.GaugeValue,
		float64(stats.Power.VoltagePhase2),
		"2",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerVoltagePhase,
		prometheus.GaugeValue,
		float64(stats.Power.VoltagePhase3),
		"3",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerCurrentPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase1),
		"1",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerCurrentPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase2),
		"2",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerCurrentPhase,
		prometheus.GaugeValue,
		float64(stats.Power.ConsumedPhase3),
		"3",
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedHour,
		prometheus.GaugeValue,
		stats.Power.ConsumedHour,
	)
	metrics <- prometheus.MustNewConstMetric(
		descPowerConsumedNet,
		prometheus.GaugeValue,
		stats.Power.ConsumedNet,
	)
	log.Println("⚙️ Collected some metrics")
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
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

func fetchSystemData() Stats {
	host := os.Getenv("SGPE_HOST")
	if host == "" {
		log.Fatal("💥 SGPE_HOST not set")
	}

	var apiResponse *apiResponse
	getDataFromApi(fmt.Sprintf(gatewayApiURL, host), &apiResponse)

	stats := Stats{
		Firmware: firmware{
			Running:         getInt(apiResponse.FirmwareRunning),
			Available:       getInt(apiResponse.FirmwareAvailable),
			UpdateAvailable: getBool(apiResponse.FirmwareUpdateAvailable),
		},
		Gas: gas{
			Consumed:     getFloat(apiResponse.GasDelivered),
			ConsumedHour: getFloat(apiResponse.GasDeliveredHour),
		},
		Power: power{
			Tariff:               getInt(apiResponse.ElectricityTariff),
			ConsumedTotalTariff1: getFloat(apiResponse.EnergyDeliveredTariff1),
			ProducedTotalTariff1: getFloat(apiResponse.EnergyReturnedTariff1),
			ConsumedTotalTariff2: getFloat(apiResponse.EnergyDeliveredTariff2),
			ProducedTotalTariff2: getFloat(apiResponse.EnergyReturnedTariff2),
			ConsumedCurrent:      getInt(apiResponse.PowerDeliveredTotal),
			ProducedCurrent:      getInt(apiResponse.PowerReturnedTotal),
			ConsumedPhase1:       getInt(apiResponse.PowerDeliveredL1),
			ConsumedPhase2:       getInt(apiResponse.PowerDeliveredL2),
			ConsumedPhase3:       getInt(apiResponse.PowerDeliveredL3),
			ProducedPhase1:       getInt(apiResponse.PowerReturnedL1),
			ProducedPhase2:       getInt(apiResponse.PowerReturnedL2),
			ProducedPhase3:       getInt(apiResponse.PowerReturnedL3),
			VoltagePhase1:        getInt(apiResponse.VoltageL1),
			VoltagePhase2:        getInt(apiResponse.VoltageL2),
			VoltagePhase3:        getInt(apiResponse.VoltageL3),
			CurrentPhase1:        getInt(apiResponse.CurrentL1),
			CurrentPhase2:        getInt(apiResponse.CurrentL2),
			CurrentPhase3:        getInt(apiResponse.CurrentL3),
			ConsumedHour:         getFloat(apiResponse.PowerDeliveredHour),
			ConsumedNet:          getFloat(apiResponse.PowerDeliveredNetto),
		},
	}

	return stats
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
