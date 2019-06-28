package shactions

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (sh *SmartHomeActions) serve(r *http.Request, request assistantRequest) interface{} {
	fulfiller, err := sh.Request(r, request.RequestID)
	if err != nil {
		return err
	}

	input := request.input()
	log.Printf("got request `%v`: %+v", request.RequestID, input)

	switch input.Intent {
	case "action.devices.SYNC":
		user, devices, err := fulfiller.Sync()
		if err != nil {
			return err
		}

		return struct {
			AgentUserID string   `json:"agentUserId"`
			Devices     []Device `json:"devices"`
		}{
			AgentUserID: user,
			Devices:     devices,
		}

	case "action.devices.QUERY":
		states, err := fulfiller.Query(input.Payload.Devices)
		if err != nil {
			return err
		} else if len(states) != len(input.Payload.Devices) {
			return errors.New("query response has invalid length")
		}

		// convert lists back to {deviceId: {...}, deviceId: {...}, ...}
		devices := make(map[string]map[string]interface{})
		for i, s := range states {
			device := input.Payload.Devices[i]
			if _, ok := devices[device.ID]; ok {
				return errors.New("duplicate reponse from Query")
			}
			devices[device.ID] = s.Extract()
		}

		return struct {
			Devices map[string]map[string]interface{} `json:"devices"`
		}{
			Devices: devices,
		}

	case "action.devices.EXECUTE":
		var rg resultGroup
		for _, command := range input.Payload.Commands {
			status, err := fulfiller.Exec(command.Devices, command.Executions)
			if err != nil {
				return err // fail early
			} else if len(status) != len(command.Devices) {
				return errors.New("exec response has invalid length")
			}

			for i, s := range status {
				device := command.Devices[i]
				seen := rg.add(device.ID, s)

				// sanity-check we only get one command-per-device
				// (docs are very vague, but seem to imply this)
				if seen {
					return fmt.Errorf("got duplicate device command: %v", device.ID)
				}
			}
		}

		return struct {
			Commands []commandResult `json:"commands"`
		}{
			Commands: rg.out(),
		}

	case "action.devices.DISCONNECT":
		return fulfiller.Disconnect()
	}

	return ErrNotSupported
}

func (sh *SmartHomeActions) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request assistantRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	payload := sh.serve(r, request)
	switch v := payload.(type) {
	case ErrorCode:
		// nb. earlier than 'error', more specific
		payload = struct {
			ErrorCode   string `json:"errorCode,omitempty"`
			DebugString string `json:"debugString,omitempty"`
		}{
			ErrorCode: v.Error(),
		}

	case error:
		// an actual error
		log.Printf("error: %v", v)
		http.Error(w, "", http.StatusInternalServerError)

	default:
		// do nothing, serve as JSON
	}

	response := struct {
		RequestID string      `json:"requestId"`
		Payload   interface{} `json:"payload"`
	}{
		RequestID: request.RequestID,
		Payload:   payload,
	}
	log.Printf("replying with payload: %+v", response.Payload)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("failed to encode output JSON: %+v", response)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
