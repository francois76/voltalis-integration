package transform

import (
	"github.com/francois76/voltalis-integration/voltalis/internal/api"
)

func syncPrograms(apiClient *api.Client) ([]string, error) {
	programs, err := apiClient.GetPrograms()
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, program := range programs {
		result = append(result, program.Name)
	}
	return result, nil
}
