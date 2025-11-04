package domain

type Node struct {
	role       NodeRole
	masterAddr string
}

func (n *Node) MasterAddr() string {
	return n.masterAddr
}

func (n *Node) SetMasterAddr(masterAddr string) {
	n.masterAddr = masterAddr
}

func (n *Node) Role() NodeRole {
	return n.role
}
