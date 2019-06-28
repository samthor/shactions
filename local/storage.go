package local

type localStorage struct {
	users   map[string]bool
	devices map[string]Device
}

type Storage interface {
}
