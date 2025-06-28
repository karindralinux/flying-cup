package port

import (
	"fmt"
	"net"
	"sync"
)

type PortManager struct {
	mu         sync.Mutex
	used       map[int]bool
	startRange int
	endRange   int
}

func NewPortManager(startPort, endPort int) *PortManager {
	return &PortManager{
		used:       make(map[int]bool),
		startRange: startPort,
		endRange:   endPort,
	}
}

// isPortAvailable checks if a port is actually available on the system (without locking)
func (p *PortManager) isPortAvailable(port int) bool {
	// Check map first (fast)
	if p.used[port] {
		return false
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// IsPortInUse checks if a specific port is currently in use
func (pm *PortManager) IsPortInUse(port int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.used[port]
}

// GetAvailablePort finds and returns an available port in the specified range
func (p *PortManager) GetAvailablePort() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.startRange; port <= p.endRange; port++ {
		// Check if port is already used in our map
		if p.used[port] {
			continue
		}

		// Check if port is available on the system
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue // Port is in use by another process
		}
		ln.Close()

		// Port is available, mark it as used
		p.used[port] = true
		return port, nil
	}

	return 0, fmt.Errorf("no available ports in the range %d-%d", p.startRange, p.endRange)
}

// ReleasePort marks a port as available for future use
func (p *PortManager) ReleasePort(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.used, port)
}
