package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	sh "github.com/samthor/shactions"
)

const (
	homegraphNotificationURL = "https://homegraph.googleapis.com/v1/devices:reportStateAndNotification"
)

type managerFulfiller struct {
	m       *Manager
	user    string // the agentUserId possibly used in response
	request string // the requestId from request
}

func (mf *managerFulfiller) Sync() (string, []sh.Device, error) {
	mf.m.lock.Lock()
	defer mf.m.lock.Unlock()

	// TODO: store mf.user, implicit "connect"

	devices := make([]sh.Device, 0, len(mf.m.devices))
	for _, ld := range mf.m.devices {
		ad := ld.ActionsDevice()
		if ad.ID != ld.ID() {
			// should never happen
			return "", nil, errors.New("ID mismatch for ID vs ActionsDevice")
		}
		devices = append(devices, ad)
	}

	return mf.user, devices, nil
}

func (mf *managerFulfiller) Query(keys []sh.DeviceKey) ([]sh.States, error) {
	return mf.m.op(keys, false, nil)
}

func (mf *managerFulfiller) Exec(keys []sh.DeviceKey, exec []sh.Exec) ([]sh.States, error) {
	return mf.m.op(keys, true, exec)
}

func (mf *managerFulfiller) Disconnect() error {
	// TODO: remove mf.user, implicit "connect"
	return nil
}

// Manager is a helper for managing simple Google Smart Home integrations.
type Manager struct {
	config  ManagerConfig
	actions *sh.SmartHomeActions

	lock    sync.Mutex // protects devices
	devices map[string]Device
}

// op wraps Exec or Query, as their return types are the same.
func (m *Manager) op(keys []sh.DeviceKey, isExec bool, exec []sh.Exec) ([]sh.States, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	states := make([]sh.States, 0, len(keys))
	for _, device := range keys {
		var err error
		var s *sh.States

		d := m.devices[device.ID]
		if d != nil {
			if isExec {
				s, err = d.Exec(exec)
			} else {
				s, err = d.Query()
			}
			if errorCode, ok := err.(sh.ErrorCode); ok {
				s = &sh.States{ErrorCode: errorCode}
			} else if err != nil {
				return nil, err // real error
			}
		}
		if s == nil {
			s = &sh.States{ErrorCode: sh.ErrDeviceNotFound}
		}

		states = append(states, *s)
	}
	return states, nil
}

// Actions returns the SmartHomeActions instance provided by Manager.
func (m *Manager) Actions() *sh.SmartHomeActions {
	return m.actions
}

// AgentUser calls the configured AgentUser or a sensible default.
func (m *Manager) AgentUser(auth string) (string, error) {
	if m.config.AgentUser != nil {
		return m.config.AgentUser(auth)
	}

	// otherwise, literally use the part after Bearer as the agentUserId
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", sh.ErrAuthFailure
	}

	user := strings.TrimSpace(auth[len("Bearer "):])
	if user == "" {
		return "", sh.ErrAuthFailure
	}

	return user, nil
}

func (m *Manager) request(r *http.Request, request string) (sh.RequestFulfiller, error) {
	auth := r.Header.Get("Authorization")
	user, err := m.AgentUser(auth)
	if err != nil {
		return nil, err
	}
	return &managerFulfiller{m: m, user: user, request: request}, nil
}

// NewManager starts a new Manager.
func NewManager(config ManagerConfig) *Manager {
	m := &Manager{
		config:  config,
		devices: make(map[string]Device),
	}
	m.actions = &sh.SmartHomeActions{Request: m.request}

	return m
}

// Seen is called when a new device is spotted on the network.
func (m *Manager) Seen(ld Device, timeout time.Duration) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	id := ld.ID()
	if id == "" {
		return fmt.Errorf("invalid ID for device: %+v", ld)
	}

	_, existed := m.devices[id]
	m.devices[id] = ld
	if !existed {
		// TODO: send requestSync request to known agentUserIDs
		log.Printf("seen new device: %v", id)
	}

	// TODO: remove after ~timeout

	return nil
}

// Report reports the state of a device to Google.
func (m *Manager) Report(ld Device, states sh.States) error {
	client := m.config.ReportStateClient
	if client == nil {
		return errors.New("no client provided for Report (needs oauth)")
	}

	payload := struct {
		Request string `json:"requestId"`
		User    string `json:"agentUserId"`
		Payload struct {
			Devices struct {
				States map[string]map[string]interface{} `json:"states"`
			} `json:"devices"`
		} `json:"payload"`
	}{
		User: "sam", // FIXME FIXME FIXME
	}

	payload.Payload.Devices.States = map[string]map[string]interface{}{
		ld.ID(): states.Extract(),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	resp, err := client.Post(homegraphNotificationURL, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("homegraph response:\n%v", string(body))
		return fmt.Errorf("got non-200 from homegraph: %v", resp.StatusCode)
	}

	return nil
}
