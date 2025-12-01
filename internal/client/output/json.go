package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONResponse is the standard JSON output format
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// OutputJSON prints data in JSON format
func OutputJSON(data interface{}, err error) {
	response := JSONResponse{
		Success: err == nil,
		Data:    data,
	}

	if err != nil {
		response.Error = err.Error()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if encodeErr := encoder.Encode(response); encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", encodeErr)
		os.Exit(1)
	}
}

// MustMarshalJSON marshals data to JSON or panics
func MustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}
