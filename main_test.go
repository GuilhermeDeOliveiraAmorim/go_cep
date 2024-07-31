package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var viacepURL = "https://viacep.com.br/ws/"

func TestGetLocationByCEP_InvalidCEP(t *testing.T) {
	invalidCEP := "123456789"

	city, statusCode, err := getLocationByCEP(invalidCEP)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, statusCode)
	assert.Empty(t, city)
}

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		cep      string
		expected bool
	}{
		{"12345-678", true},
		{"12345678", true},
		{"1234-5678", false},
		{"1234567a", false},
	}

	for _, test := range tests {
		if result := isValidCEP(test.cep); result != test.expected {
			t.Errorf("isValidCEP(%s) = %v; want %v", test.cep, result, test.expected)
		}
	}
}

func TestGetLocationByCEP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"localidade": "São Paulo"}`))
	}))
	defer server.Close()

	oldURL := "https://viacep.com.br/ws/"
	defer func() { viacepURL = oldURL }()
	viacepURL = server.URL

	city, statusCode, err := getLocationByCEP("01001-000")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, statusCode)
	}

	expectedCity := "São Paulo"
	if city != expectedCity {
		t.Errorf("Expected city %s, got %s", expectedCity, city)
	}
}

func TestGetWeatherByCity(t *testing.T) {
	client := createHTTPClient()
	apiKey := "87022f0c0e0d4335a1d182957242207"
	encodedCity := url.QueryEscape("São Paulo")
	url := "https://api.weatherapi.com/v1/current.json?key=" + apiKey + "&q=" + encodedCity
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	current, ok := data["current"].(map[string]interface{})
	if !ok {
		t.Fatalf("Unexpected error: %v", err)
	}

	tempNow, ok := current["temp_c"].(float64)
	if !ok {
		t.Fatalf("Unexpected error: %v", err)
	}

	tempC, err := getWeatherByCity("São Paulo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedTempC := tempNow
	if tempC != expectedTempC {
		t.Errorf("Expected temperature %f, got %f", expectedTempC, tempC)
	}
}

func TestConvertTemperature(t *testing.T) {
	tempC := 25.0
	expectedF := 77.0
	expectedK := 298.15

	tempF, tempK := convertTemperature(tempC)
	if tempF != expectedF {
		t.Errorf("Expected tempF %f, got %f", expectedF, tempF)
	}
	if tempK != expectedK {
		t.Errorf("Expected tempK %f, got %f", expectedK, tempK)
	}
}

func TestGetLocationByErrorCEP(t *testing.T) {
	tests := []struct {
		cep          string
		mockResponse string
		statusCode   int
		expectedCity string
		expectedErr  string
	}{
		{"01001000", `{"localidade": "São Paulo"}`, http.StatusOK, "São Paulo", ""},
		{"12345678", ``, http.StatusNotFound, "", "can not find zipcode"},
		{"123456789", ``, http.StatusUnprocessableEntity, "", "invalid zipcode"},
	}

	for _, test := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.mockResponse != "" {
				w.Write([]byte(test.mockResponse))
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
			}
		}))
		defer server.Close()

		viacepURL = server.URL

		city, statusCode, err := getLocationByCEP(test.cep)
		if err != nil && err.Error() != test.expectedErr {
			t.Errorf("Expected error %v, got %v", test.expectedErr, err)
		}

		if statusCode != test.statusCode {
			t.Errorf("Expected status code %d, got %d", test.statusCode, statusCode)
		}

		if city != test.expectedCity {
			t.Errorf("Expected city %s, got %s", test.expectedCity, city)
		}
	}
}
