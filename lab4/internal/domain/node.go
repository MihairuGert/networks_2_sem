package domain

type Node struct {
	role NodeRole
}

func (n Node) Role() NodeRole {
	return n.role
}
