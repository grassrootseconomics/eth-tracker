package event

import "encoding/json"

type (
	Event struct {
		Block           uint64         `json:"block"`
		ContractAddress string         `json:"contractAddress"`
		Success         bool           `json:"success"`
		Timestamp       uint64         `json:"timestamp"`
		TxHash          string         `json:"transactionHash"`
		TxType          string         `json:"transactionType"`
		Payload         map[string]any `json:"payload"`
	}
)

func (e Event) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return jsonData, err
}
