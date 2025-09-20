package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Options struct {
	MqttURL          string `json:"mqtt_url"`
	MqttPassword     string `json:"mqtt_password"`
	VoltalisLogin    string `json:"voltalis_login"`
	VoltalisPassword string `json:"voltalis_password"`
}

func LoadOptions() (*Options, error) {
	// Récupérer le chemin du fichier depuis la variable d'env
	path := os.Getenv("OPTIONS_FILE")
	if path == "" {
		return nil, fmt.Errorf("la variable d'environnement OPTIONS_FILE n'est pas définie")
	}

	// Lire le fichier
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le fichier %s: %w", path, err)
	}

	// Unmarshal JSON vers la struct
	var opts Options
	if err := json.Unmarshal(data, &opts); err != nil {
		return nil, fmt.Errorf("erreur de parsing JSON: %w", err)
	}

	return &opts, nil
}
