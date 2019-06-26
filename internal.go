package shactions

// assistantRequest is the top-level payload received for Smart Home requests.
type assistantRequest struct {
	RequestID string           `json:"requestId"`
	Inputs    []assistantInput `json:"inputs"`
}

// input returns the single input as part of an assistantRequest.
func (ahr *assistantRequest) input() assistantInput {
	if ahr != nil && len(ahr.Inputs) == 1 {
		return ahr.Inputs[0]
	}
	return assistantInput{}
}

// assistantInput is a specific command as part of assistantHomeRequest.
type assistantInput struct {
	Intent  string `json:"intent"`
	Payload struct {
		Commands []assistantCommand `json:"commands"`
		Devices  []DeviceKey        `json:"devices"`
	} `json:"payload"`
}

// assistantCommand groups a number of executions to be run on a number of devices.
type assistantCommand struct {
	Devices    []DeviceKey `json:"devices"`
	Executions []Exec      `json:"execution"`
}

// assistantResponse contains the top-level response type.
type assistantResponse struct {
	ErrorCode   string `json:"errorCode,omitempty"`
	DebugString string `json:"debugString,omitempty"`
}

// commandResult is used inside the response to an exec request.
type commandResult struct {
	Devices   []string               `json:"ids"`
	Status    string                 `json:"status"`
	ErrorCode string                 `json:"errorCode"`
	States    map[string]interface{} `json:"states"`
}
