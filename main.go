package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func createHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr, Timeout: 10 * time.Second}
}

func getLocationByCEP(cep string) (string, error) {
	client := createHTTPClient()
	resp, err := client.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CEP não encontrado")
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	city, ok := data["localidade"]
	if !ok {
		return "", fmt.Errorf("CEP não encontrado")
	}

	return city, nil
}

func getWeatherByCity(city string) (float64, error) {
	client := createHTTPClient()
	apiKey := "87022f0c0e0d4335a1d182957242207"
	resp, err := client.Get(fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("cidade não encontrada")
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	current, ok := data["current"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("dados do clima não encontrados")
	}

	tempC, ok := current["temp_c"].(float64)
	if !ok {
		return 0, fmt.Errorf("temperatura não encontrada")
	}

	return tempC, nil
}

func convertTemperature(tempC float64) (float64, float64) {
	tempF := tempC*1.8 + 32

	tempK := tempC + 273.15
	tempKStr := fmt.Sprintf("%.2f", tempK)
	tempKFloat, err := strconv.ParseFloat(tempKStr, 64)
	if err != nil {
		fmt.Println("Erro ao converter temperatura para float64:", err)
	}

	return tempF, tempKFloat
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	cep := strings.TrimPrefix(r.URL.Path, "/weather/")

	city, err := getLocationByCEP(cep)
	if err != nil {
		http.Error(w, "can not find zipcode (location) "+err.Error(), http.StatusNotFound)
		return
	}

	tempC, err := getWeatherByCity(city)
	if err != nil {
		http.Error(w, "can not find zipcode (weather) "+err.Error(), http.StatusNotFound)
		return
	}

	tempF, tempK := convertTemperature(tempC)

	response := map[string]float64{
		"temp_C": tempC,
		"temp_F": tempF,
		"temp_K": tempK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/weather/", weatherHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Olá, mundo!"))
	})
	http.ListenAndServe(":8080", nil)
}
