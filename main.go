package main

import (
	"os"
	"os/exec"
	"sync"

	"github.com/rs/zerolog/log"
)

func runMicroservices(command string) {
	cmd := exec.Command("go", "run", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal().Err(err).Msgf("Error running microservices %s", command)
	}
}

func main() {
	var wg sync.WaitGroup

	microservices := []string{
		"internal/app/auth/main.go",
	}

	for _, ms := range microservices {
		wg.Add(1)
		go func(microservice string) {
			defer wg.Done()
			runMicroservices(microservice)
		}(ms)
	}

	wg.Wait()
}
