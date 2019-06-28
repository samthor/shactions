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

// commandResult is used inside the response to an exec request.
type commandResult struct {
	Devices   []string               `json:"ids"`
	Status    string                 `json:"status"`
	ErrorCode string                 `json:"errorCode"`
	States    map[string]interface{} `json:"states"`
}

// resultKey is used as part of resultGroup.
type resultKey struct {
	Status
	ErrorCode
}

// resultValue is used as part of resultGroup.
type resultValue struct {
	devices []string
	states  []States
}

// resultGroup helps to group results from an exec request.
type resultGroup struct {
	m    map[resultKey]resultValue
	seen map[string]bool
}

func (rg *resultGroup) out() []commandResult {
	out := make([]commandResult, 0, len(rg.m))
	for k, v := range rg.m {
		out = append(out, commandResult{
			Devices:   v.devices,
			Status:    k.Status.String(),
			ErrorCode: k.ErrorCode.Error(),
			States:    mergeStates(v.states),
		})
	}
	return out
}

func (rg *resultGroup) add(device string, s States) bool {
	if rg.m == nil {
		rg.m = make(map[resultKey]resultValue)
		rg.seen = make(map[string]bool)
	}

	k := resultKey{s.Status, s.ErrorCode}
	v := rg.m[k]

	v.devices = append(v.devices, device)

	// reset parts of States which aren't useful to Exec
	s.Status = Unknown
	s.ErrorCode = ErrorCode{nil}
	v.states = append(v.states, s)

	rg.m[k] = v

	if rg.seen[device] {
		return true
	}
	rg.seen[device] = true
	return false
}

func mergeStates(states []States) map[string]interface{} {
	if len(states) == 0 {
		return nil
	}
	out := states[0].Extract()
	for _, s := range states[1:] {
		out = intersectMap(out, s.Extract())
	}
	return out
}
