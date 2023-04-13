/*
 * @Author: jiale_quan jiale_quan@ustc.edu
 * @Date: 2023-04-13 20:01:12
 * @LastEditTime: 2023-04-13 20:31:49
 * @Description:
 * Copyright © jiale_quan, All Rights Reserved
 */
package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type JLRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func (d *JLRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *JLRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}

	servers := strings.Split(resp.Header.Get("X-jlrpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))

	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *JLRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}

func (d *JLRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}

func NewJLRegistryDiscovery(registryAddr string, timeout time.Duration) *JLRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &JLRegistryDiscovery{
		MultiServersDiscovery: NewMultiserverDiscovery(make([]string, 0)),
		registry:              registryAddr,
		timeout:               timeout,
	}
	return d
}
