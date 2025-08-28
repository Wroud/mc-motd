package server

import (
	"context"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Server struct {
	ctx         context.Context
	config      *Config
	connector   *Connector
	motdManager *MOTDManager
	doneChan    chan struct{}
}

func NewServer(ctx context.Context, config *Config) (*Server, error) {
	motdManager := NewMOTDManager(&config.ServerStatus)
	connector := NewConnector(ctx, config, motdManager)

	if config.Webhook.Url != "" {
		logrus.WithField("url", config.Webhook.Url).
			WithField("require-user", config.Webhook.RequireUser).
			Info("Using webhook for connection status notifications")
		connector.UseConnectionNotifier(
			NewWebhookNotifier(config.Webhook.Url, config.Webhook.RequireUser))
	}

	return &Server{
		ctx:         ctx,
		config:      config,
		connector:   connector,
		motdManager: motdManager,
		doneChan:    make(chan struct{}),
	}, nil
}

// Done provides a channel notified when the server has closed all connections, etc
func (s *Server) Done() <-chan struct{} {
	return s.doneChan
}

func (s *Server) notifyDone() {
	s.doneChan <- struct{}{}
}

// AcceptConnection provides a way to externally supply a connection to consume
// Note that this will skip rate limiting.
func (s *Server) AcceptConnection(conn net.Conn) {
	s.connector.AcceptConnection(conn)
}

// Run will run the server until the context is done or a fatal error occurs, so this should be
// in a go routine.
func (s *Server) Run() {
	defer s.motdManager.Close() // Clean up MOTD manager when server stops

	err := s.connector.StartAcceptingConnections(
		net.JoinHostPort("", strconv.Itoa(s.config.Port)),
	)
	if err != nil {
		logrus.WithError(err).Error("Could not start accepting connections")
		s.notifyDone()
		return
	}

	<-s.ctx.Done()
	logrus.Info("Stopped")
	s.notifyDone()
}
