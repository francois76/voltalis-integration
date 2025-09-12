package main

import (
	"fmt"
	"time"

	"github.com/tonuser/voltalis-integration/voltalis/internal/ha"
)

func main() {
	fmt.Println("Voltalis add-on starting...")

	client := ha.NewClient()

	i := 0
	for {
		state := fmt.Sprintf("update-%d", i)
		fmt.Println("Publishing state:", state)
		client.PublishState("sensor.voltalis_status", state, map[string]any{
			"friendly_name": "Voltalis Status",
		})
		i++
		time.Sleep(15 * time.Second)
	}
}
