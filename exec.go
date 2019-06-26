package shactions

type resultKey struct {
	Status
	ErrorCode
}

type resultValue struct {
	devices []string
	states  []States
}

type resultGroup struct {
	m map[resultKey]resultValue
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

func (rg *resultGroup) add(device string, s ExecStatus) {
	if rg.m == nil {
		rg.m = make(map[resultKey]resultValue)
	}

	k := resultKey{s.Status, s.ErrorCode}
	v := rg.m[k]

	// TODO: check for duplicates
	v.devices = append(v.devices, device)
	v.states = append(v.states, s.States)

	rg.m[k] = v
}

func mergeStates(states []States) map[string]interface{} {
	if len(states) == 0 {
		return nil
	}
	out := states[0].extract()
	for _, s := range states[1:] {
		out = intersectMap(out, s.extract())
	}
	return out
}
