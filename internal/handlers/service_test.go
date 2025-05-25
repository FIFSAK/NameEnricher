package handlers

import (
	"NameEnricher/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.Init()

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestAgeFromExternalApi(t *testing.T) {
	agifyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			t.Error("Name parameter is missing")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var age int
		switch name {
		case "John":
			age = 35
		case "Mary":
			age = 28
		case "ErrorCase":
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
			age = 30
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": name,
			"age":  age,
		})
	}))
	defer agifyServer.Close()

	//originalUrl := "https://api.agify.io/"

	tests := []struct {
		name       string
		personName string
		wantAge    int
		wantErr    bool
	}{
		{"Valid name John", "John", 35, false},
		{"Valid name Mary", "Mary", 28, false},
		{"Default age", "Unknown", 30, false},
		{"Error case", "ErrorCase", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Временно подменяем URL для тестирования
			http.DefaultClient.Transport = &mockTransport{URL: agifyServer.URL}

			age, err := ageFromExternalApi(tt.personName)

			// Проверяем результаты
			if (err != nil) != tt.wantErr {
				t.Errorf("ageFromExternalApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && age != tt.wantAge {
				t.Errorf("ageFromExternalApi() = %v, want %v", age, tt.wantAge)
			}
		})
	}

	http.DefaultClient.Transport = nil
}

func TestGenderFromExternalApi(t *testing.T) {
	genderizeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			t.Error("Name parameter is missing")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var gender string
		switch name {
		case "John":
			gender = "male"
		case "Mary":
			gender = "female"
		case "ErrorCase":
			w.WriteHeader(http.StatusInternalServerError)
			return
		case "EmptyGender":
			gender = ""
		default:
			gender = "unknown"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":   name,
			"gender": gender,
		})
	}))
	defer genderizeServer.Close()

	tests := []struct {
		name       string
		personName string
		wantGender string
		wantErr    bool
	}{
		{"Valid name John", "John", "male", false},
		{"Valid name Mary", "Mary", "female", false},
		{"Unknown gender", "Unknown", "unknown", false},
		{"Empty gender", "EmptyGender", "", false},
		{"Error case", "ErrorCase", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultClient.Transport = &mockTransport{URL: genderizeServer.URL}

			gender, err := genderFromExternalApi(tt.personName)

			if (err != nil) != tt.wantErr {
				t.Errorf("genderFromExternalApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gender != tt.wantGender {
				t.Errorf("genderFromExternalApi() = %v, want %v", gender, tt.wantGender)
			}
		})
	}

	http.DefaultClient.Transport = nil
}

func TestNationalityFromExternalApi(t *testing.T) {
	nationalizeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			t.Error("Name parameter is missing")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var response map[string]interface{}
		switch name {
		case "John":
			response = map[string]interface{}{
				"name": name,
				"country": []map[string]interface{}{
					{"country_id": "US", "probability": 0.8},
					{"country_id": "GB", "probability": 0.1},
				},
			}
		case "Boris":
			response = map[string]interface{}{
				"name": name,
				"country": []map[string]interface{}{
					{"country_id": "RU", "probability": 0.7},
					{"country_id": "BG", "probability": 0.2},
				},
			}
		case "ErrorCase":
			w.WriteHeader(http.StatusInternalServerError)
			return
		case "EmptyCountries":
			response = map[string]interface{}{
				"name":    name,
				"country": []map[string]interface{}{},
			}
		default:
			response = map[string]interface{}{
				"name": name,
				"country": []map[string]interface{}{
					{"country_id": "XX", "probability": 0.5},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer nationalizeServer.Close()

	tests := []struct {
		name            string
		personName      string
		wantNationality string
		wantErr         bool
	}{
		{"Valid name John", "John", "US", false},
		{"Valid name Boris", "Boris", "RU", false},
		{"Default nationality", "Unknown", "XX", false},
		{"Empty countries list", "EmptyCountries", "", true},
		{"Error case", "ErrorCase", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultClient.Transport = &mockTransport{URL: nationalizeServer.URL}

			nationality, err := nationalityFromExternalApi(tt.personName)

			if (err != nil) != tt.wantErr {
				t.Errorf("nationalityFromExternalApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && nationality != tt.wantNationality {
				t.Errorf("nationalityFromExternalApi() = %v, want %v", nationality, tt.wantNationality)
			}
		})
	}

	http.DefaultClient.Transport = nil
}

type mockTransport struct {
	URL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := m.URL + req.URL.Path + "?" + req.URL.RawQuery
	newReq, err := http.NewRequest(req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}

	for k, vv := range req.Header {
		for _, v := range vv {
			newReq.Header.Add(k, v)
		}
	}

	return http.DefaultTransport.RoundTrip(newReq)
}
