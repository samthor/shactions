package shactions

import (
	"net/http"
)

// DeviceKey identifies a device by its ID and optional custom key/value pairs.
type DeviceKey struct {
	ID         string                 `json:"id"`
	CustomData map[string]interface{} `json:"customData,omitempty"`
}

// DeviceInfo describes a device's manufacturer, model etc.
type DeviceInfo struct {
	Manufacturer    string `json:"manufacturer,omitempty"`
	Model           string `json:"model,omitempty"`
	HardwareVersion string `json:"hwVersion,omitempty"`
	SoftwareVersion string `json:"swVersion,omitempty"`
}

// Device describes a device which can be controlled.
type Device struct {
	DeviceKey

	Type            string                 `json:"type"`
	Traits          []string               `json:"traits"`          // the possible ways this can be controleld
	WillReportState bool                   `json:"willReportState"` // true is real-time, false is polling
	RoomHint        string                 `json:"roomHint,omitempty"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`
	DeviceInfo      *DeviceInfo            `json:"deviceInfo,omitempty"`

	Name struct {
		Name         string   `json:"name"`
		DefaultNames []string `json:"defaultNames,omitempty"`
		Nicknames    []string `json:"nicknames,omitempty"`
	} `json:"name"`
}

// Exec identifies a general user request.
type Exec struct {
	Task   string                 `json:"command"`
	Params map[string]interface{} `json:"params"`
}

// SmartHomeActions is...
type SmartHomeActions struct {
	Request func(*http.Request, string) (RequestFulfiller, error)
}

// RequestFulfiller fulfills a single request (potentially in a series).
type RequestFulfiller interface {
	Sync() (string, []Device, error)
	Query([]DeviceKey) ([]States, error)
	Exec([]DeviceKey, []Exec) ([]ExecStatus, error)
	Disconnect() error
}

// ExecStatus is the status of executing many commands on one device.
type ExecStatus struct {
	Status    Status
	ErrorCode ErrorCode // subtype of Error
	States    States
}

// States is a helper that wraps a Map plus some known, fixed keys.
type States struct {
	Online      bool
	ErrorCode   ErrorCode
	DebugString string
	m           map[string]interface{}
}

// Set is a convenience helper to set a map value.
func (s *States) Set(key string, value interface{}) {
	if s.m == nil {
		s.m = make(map[string]interface{})
	}
	s.m[key] = value
}

// extract converts this States into a single map.
func (s States) extract() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range s.m {
		out[k] = v
	}

	// write known keys
	out["online"] = s.Online
	if s.ErrorCode.error != nil {
		out["errorCode"] = s.ErrorCode.error
	}
	if s.DebugString != "" {
		out["debugString"] = s.DebugString
	}

	return out
}

// Status is a status returned by this API.
type Status int

func (status Status) String() string {
	switch status {
	case Success:
		return "SUCCESS"
	case Pending:
		return "PENDING"
	case Offline:
		return "OFFLINE"
	case Error:
		return "ERROR"
	}
	return "" // includes Unknown
}

const (
	// Unknown is the default, zero state of Status.
	Unknown Status = iota

	// Success confirms that the command(s) has/have succeeded.
	Success

	// Pending indicates that the command(s) are enqueued but expected to succeed.
	Pending

	// Offline indicates that the target devices(s) are in an offline state or unreachable.
	Offline

	// Error requires that the ErrorCode field is also set.
	Error
)
