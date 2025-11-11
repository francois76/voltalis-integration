package transform

import (
	"github.com/francois76/voltalis-integration/voltalis/internal/api"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func syncPrograms(controller *mqtt.Controller, apiClient *api.Client) error {
	programs, err := apiClient.GetPrograms()
	if err != nil {
		return err
	}
	result := []string{}
	for _, program := range programs {
		result = append(result, program.Name)
	}
	controller.AddSelectProgram(result...)
	return nil
}
