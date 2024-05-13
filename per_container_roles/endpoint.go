package per_container_roles

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Endpoint struct {
	PortNum          int
	Server           *http.Server
	NetworkID        string
	ByContainer      map[string]*ContainerWithCreds
	ByContainerMutex sync.Mutex
}

func (e *Endpoint) RemoveAllContainers() {
	e.ByContainerMutex.Lock()
	e.ByContainer = make(map[string]*ContainerWithCreds)
	e.ByContainerMutex.Unlock()
}

func (e *Endpoint) AddContainer(id string, role string, ip string) error {
	// e.ByContainer[id] = ContainerWithCreds{IPAddress: ip, Role: role}
	roleARN, err := arn.Parse(role)
	if err != nil {
		return fmt.Errorf("container: %s: could not parse role ARN %s: %s", id, role, err)
	}

	roleARNResourceParts := strings.Split(roleARN.Resource, "/")

	roleName := roleARNResourceParts[len(roleARNResourceParts)-1]

	e.ByContainerMutex.Lock()
	e.ByContainer[id] = &ContainerWithCreds{
		IPAddress:       ip,
		RoleARN:         role,
		RoleName:        roleName,
		RoleSessionName: id,
		Creds: RefreshableCred{
			Expiration:  time.Now(),
			LastUpdated: time.Now(),
			Code:        REFRESHABLE_CRED_CODE,
			Type:        REFRESHABLE_CRED_TYPE,
		},
	}
	e.ByContainerMutex.Unlock()

	return nil
}

func (e *Endpoint) RemoveContainer(id string) {
	e.ByContainerMutex.Lock()
	delete(e.ByContainer, id)
	e.ByContainerMutex.Unlock()
}

func (e *Endpoint) CredsByIP(ip string) (*ContainerWithCreds, bool) {
	for _, container := range e.ByContainer {
		if container.IPAddress == ip {
			return container, true
		}
	}
	return nil, false
}

func (e *Endpoint) LoadContainersFromDocker(cli *client.Client, ctx context.Context) error {
	// Mine:
	network, err := cli.NetworkInspect(ctx, e.NetworkID, types.NetworkInspectOptions{})
	if err != nil {
		return fmt.Errorf("could not inspect network: %s", err.Error())
	}

	e.RemoveAllContainers()

	for containerId, container := range network.Containers {
		inspectedContainer, err := cli.ContainerInspect(ctx, containerId)

		if err != nil {
			log.Println("Error inspecting container:", containerId, err)
			continue
		}

		roleLabel, ok := inspectedContainer.Config.Labels["per-container-role"]
		if !ok {
			log.Println("Container does not have a per-container-role label")
			continue
		}

		network, ok := inspectedContainer.NetworkSettings.Networks[e.NetworkID]
		if !ok {
			log.Println("Container does not have the network:", e.NetworkID)
			continue
		}

		log.Println("Found container:", containerId, container.Name, network.IPAddress, roleLabel)

		err = e.AddContainer(containerId, roleLabel, network.IPAddress)
		if err != nil {
			log.Println("Error adding container:", containerId, err)
		}

	}

	return nil
}

func (e *Endpoint) MonitorNetworkEvents(cli *client.Client, ctx context.Context) {
	// TODO: filter by the deisred network so we don't have to inspect every container
	eventsMessages, eventsErrors := cli.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "type", Value: "network"},
		),
	})

	for event := range eventsMessages {
		if event.Type == "network" && event.Action == "connect" {
			containerId := event.Actor.Attributes["container"]

			inspectedContainer, err := cli.ContainerInspect(ctx, containerId)

			if err != nil {
				log.Println("Error inspecting container:", event.Actor.Attributes["container"], err)
				continue
			}

			roleLabel, ok := inspectedContainer.Config.Labels["per-container-role"]
			if !ok {
				log.Println("Container does not have a per-container-role label. Container:", containerId)
				continue
			}

			network, ok := inspectedContainer.NetworkSettings.Networks[e.NetworkID]
			if !ok {
				log.Println("Container does not have expected network. Container:", containerId, "Network:", e.NetworkID)
				continue
			}

			log.Println("Adding container:", containerId, inspectedContainer.Name, "IP:", network.IPAddress, "Role:", roleLabel)
			e.AddContainer(containerId, roleLabel, network.IPAddress)

		} else if event.Type == "network" && event.Action == "disconnect" {
			containerId := event.Actor.Attributes["container"]
			log.Println("Container disconnected. Container:", containerId)
			e.RemoveContainer(containerId)
		}
	}

	err := <-eventsErrors
	log.Println("docker Events Error:", err)
}

func (e *Endpoint) ConfigureFromDocker(cli *client.Client, ctx context.Context) (chan string, chan error) {
	ready := make(chan string)
	errors := make(chan error)

	go func() {
		for {
			err := e.LoadContainersFromDocker(cli, ctx)
			if err != nil {
				errors <- fmt.Errorf("could not load containers: %s", err.Error())
				close(ready)
				close(errors)
				return
			}
			ready <- "loaded containers"

			e.MonitorNetworkEvents(cli, ctx)
		}
	}()

	return ready, errors
}
