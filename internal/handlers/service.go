package handlers

import (
	"NameEnricher/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

func ageFromExternalApi(name string) (int, error) {
	logger.Log.Infof("Requesting age data for name: %s", name)

	apiUrl := fmt.Sprintf("https://api.agify.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		logger.Log.Errorf("Failed to request age API: %v", err)
		return 0, fmt.Errorf("failed to request age API: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Debugf("Received response from age API with status: %s", resp.Status)

	var response struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Log.Errorf("Failed to decode age API response: %v", err)
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Log.Infof("Successfully determined age %d for name: %s", response.Age, name)
	return response.Age, nil
}

func genderFromExternalApi(name string) (string, error) {
	logger.Log.Infof("Requesting gender data for name: %s", name)

	apiUrl := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		logger.Log.Errorf("Failed to request gender API: %v", err)
		return "", fmt.Errorf("failed to request gender API: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Debugf("Received response from gender API with status: %s", resp.Status)

	var response struct {
		Gender string `json:"gender"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Log.Errorf("Failed to decode gender API response: %v", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Log.Infof("Successfully determined gender '%s' for name: %s", response.Gender, name)
	return response.Gender, nil
}

func nationalityFromExternalApi(name string) (string, error) {
	logger.Log.Infof("Requesting nationality data for name: %s", name)

	apiUrl := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		logger.Log.Errorf("Failed to request nationality API: %v", err)
		return "", fmt.Errorf("failed to request nationality API: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Debugf("Received response from nationality API with status: %s", resp.Status)

	var response struct {
		Country []struct {
			CountryId   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Log.Errorf("Failed to decode nationality API response: %v", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Country) == 0 {
		logger.Log.Warnf("No nationality data found for name: %s", name)
		return "", fmt.Errorf("country not found for name: %s", name)
	}

	var result string
	maxProbability := float64(0)
	for _, country := range response.Country {
		if country.Probability > maxProbability {
			maxProbability = country.Probability
			result = country.CountryId
		}
	}

	logger.Log.Infof("Successfully determined nationality '%s' (probability: %.2f) for name: %s",
		result, maxProbability, name)
	return result, nil
}
