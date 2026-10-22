package daemon

import "sync"

var stateCtr = &stateCounter{states: make(map[string]string)}

// stateCounter keeps track of container states for the info endpoint.
type stateCounter struct {
	mu     sync.Mutex
	states map[string]string
}

func (ctr *stateCounter) get() (running int, paused int, stopped int) {
	ctr.mu.Lock()
	defer ctr.mu.Unlock()

	states := map[string]int{
		"running": 0,
		"paused":  0,
		"stopped": 0,
	}
	for _, state := range ctr.states {
		states[state]++
	}
	return states["running"], states["paused"], states["stopped"]
}

func (ctr *stateCounter) set(id, label string) {
	ctr.mu.Lock()
	ctr.states[id] = label
	ctr.mu.Unlock()
}

func (ctr *stateCounter) del(id string) {
	ctr.mu.Lock()
	delete(ctr.states, id)
	ctr.mu.Unlock()
}
