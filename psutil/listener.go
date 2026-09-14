package psutil

type Listener struct {
	Protocol string
	Address  string
	Port     uint16
	Pid      int
	Name     string
}

func GetListeners() (listeners []*Listener, err error) {
	listeners, err = listenerList()
	if err != nil {
		return
	}

	return
}
