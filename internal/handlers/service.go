package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ageFromExternalApi(name string) (int, error) {
	apiUrl := fmt.Sprintf("https://api.agify.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return 0, fmt.Errorf("error during requesting API: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Age, nil
}

func genderFromExternalApi(name string) (string, error) {
	apiUrl := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("error during requesting API: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Gender string `json:"gender"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Gender, nil
}

func nationalityFromExternalApi(name string) (string, error) {
	apiUrl := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("error during requesting API: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Country []struct {
			CountryId   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Country) == 0 {
		return "", fmt.Errorf("nationality not found for name: : %s", name)
	}

	var result string
	maxProbability := float64(0)
	for _, country := range response.Country {
		if country.Probability > maxProbability {
			maxProbability = country.Probability
			result = country.CountryId
		}
	}
	return result, nil
}
