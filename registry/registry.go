/*
 * @Author: jiale_quan jiale_quan@ustc.edu
 * @Date: 2023-04-13 19:34:30
 * @LastEditTime: 2023-04-13 19:57:01
 * @Description:
 * Copyright © jiale_quan, All Rights Reserved
 */
package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type JLRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_jlrpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *JLRegistry {
	return &JLRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultJLRegister = New(defaultTimeout)

func (r *JLRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string

	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

/*
 * run at /_jlrpc_/registry
 */
func (r *JLRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-jlrpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-jlrpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *JLRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		s.start = time.Now()
	}
}

/*
 * HandleHTTP registers an HTTP handler for JLRegistry messages on registrtPath
 */
func (r *JLRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultJLRegister.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-jlrpc-Server", addr)

	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
