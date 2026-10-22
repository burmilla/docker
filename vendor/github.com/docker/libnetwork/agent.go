package libnetwork

// Swarm mode networking agent (networkdb gossip) is not supported in
// System Docker; these hooks are kept as no-ops for the callers.

type agent struct{}

// Task has the backend container details
type Task struct {
	Name       string
	EndpointID string
	EndpointIP string
	Info       map[string]string
}

// ServiceInfo has service specific details along with the list of backend tasks
type ServiceInfo struct {
	VIP          string
	LocalLBIndex int
	Tasks        []Task
	Ports        []string
}

func (c *controller) getAgent() *agent { return nil }

func (c *controller) agentDriverNotify(d interface{}) {}

func (c *controller) agentClose() {}

func (n *network) Services() map[string]ServiceInfo { return nil }

func (n *network) isClusterEligible() bool { return false }

func (n *network) joinCluster() error { return nil }

func (n *network) leaveCluster() error { return nil }

func (ep *endpoint) addDriverInfoToCluster() error { return nil }

func (ep *endpoint) deleteDriverInfoFromCluster() error { return nil }

func (ep *endpoint) addServiceInfoToCluster(sb *sandbox) error { return nil }

func (ep *endpoint) deleteServiceInfoFromCluster(sb *sandbox, method string) error { return nil }

func (n *network) addDriverWatches() {}

func (n *network) cancelDriverWatches() {}
