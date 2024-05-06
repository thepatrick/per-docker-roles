package per_container_roles

import "sync"

type ContainerWithCreds struct {
	IPAddress       string
	RoleARN         string
	RoleName        string
	RoleSessionName string
	Creds           RefreshableCred
	Mutex           sync.Mutex
}

func (c *ContainerWithCreds) UpdateCreds(creds RefreshableCred) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Creds = creds
}
