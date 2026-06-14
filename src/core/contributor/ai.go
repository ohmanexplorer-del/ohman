package contributor

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseTriageDecision(response string) (*TriageDecision, error) {
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("json object not found")
	}

	var decision TriageDecision
	if err := json.Unmarshal([]byte(response[start:end+1]), &decision); err != nil {
		return nil, err
	}
	if decision.Confidence < 0 {
		decision.Confidence = 0
	}
	if decision.Confidence > 1 {
		decision.Confidence = 1
	}
	return &decision, nil
}

func parsePatchDecision(response string) (*PatchDecision, error) {
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("json object not found")
	}

	var decision PatchDecision
	if err := json.Unmarshal([]byte(response[start:end+1]), &decision); err != nil {
		return nil, err
	}
	return &decision, nil
}
