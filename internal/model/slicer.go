package model

type Address struct {
	Host string
}

type ConnectedNodes struct {
	Hosts []Address
}

type DisconnectedNodes struct {
	Hosts []Address
}

type UpdateNodes struct {
	New          ConnectedNodes
	Disconnected DisconnectedNodes
}

func (un *UpdateNodes) ContainsEvents() bool {
	return len(un.New.Hosts) != 0 || len(un.Disconnected.Hosts) != 0
}
