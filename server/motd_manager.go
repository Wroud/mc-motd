package server

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type MOTDManager struct {
	mu             sync.RWMutex
	config         *ServerStatusConfig
	startingExpire time.Time
}

func NewMOTDManager(config *ServerStatusConfig) *MOTDManager {
	return &MOTDManager{
		config: config,
	}
}

func (m *MOTDManager) GetCurrentMOTD() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if time.Now().Before(m.startingExpire) {
		return m.config.StartingMOTD
	}
	return m.config.SleepingMOTD
}

func (m *MOTDManager) OnJoinAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()

	timeout := time.Duration(m.config.StartingTimeout) * time.Second
	m.startingExpire = time.Now().Add(timeout)

	logrus.WithFields(logrus.Fields{
		"timeout":   timeout,
		"expire_at": m.startingExpire,
	}).Info("Join attempt received, server showing starting MOTD")
}

func (m *MOTDManager) Close() {
}
