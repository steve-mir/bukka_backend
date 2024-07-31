package factory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//  Command pattern along with the Factory pattern to make the code more extensible and maintainable while keeping it optimal.
// This approach will allow us to easily add new services without modifying existing code, adhering to the Open/Closed Principle.

// Command interface
type Command interface {
	Execute() ([]byte, error)
}

// CommandFactory interface
type CommandFactory interface {
	CreateCommand([]byte) (Command, error)
}

func sendRequest(url string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
