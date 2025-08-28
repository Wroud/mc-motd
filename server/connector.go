package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wroud/mc-motd/mcproto"
)

const (
	handshakeTimeout = 5 * time.Second
)

var noDeadline time.Time

func NewConnector(ctx context.Context, config *Config, motdManager *MOTDManager) *Connector {

	return &Connector{
		ctx:         ctx,
		config:      config,
		motdManager: motdManager,
	}
}

type Connector struct {
	ctx                context.Context
	config             *Config
	motdManager        *MOTDManager
	state              mcproto.State
	connectionNotifier ConnectionNotifier
}

func (c *Connector) UseConnectionNotifier(notifier ConnectionNotifier) {
	c.connectionNotifier = notifier
}

func (c *Connector) StartAcceptingConnections(listenAddress string) error {
	ln, err := c.createListener(listenAddress)
	if err != nil {
		return err
	}

	go c.acceptConnections(ln)

	return nil
}

func (c *Connector) createListener(listenAddress string) (net.Listener, error) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to start listening")
		return nil, err
	}
	logrus.WithField("listenAddress", listenAddress).Info("Listening for Minecraft client connections")

	return listener, nil
}

// AcceptConnection provides a way to externally supply a connection to consume.
// Note that this will skip rate limiting.
func (c *Connector) AcceptConnection(conn net.Conn) {
	go c.HandleConnection(conn)
}

func (c *Connector) acceptConnections(ln net.Listener) {
	defer ln.Close()

	for {
		select {
		case <-c.ctx.Done():
			return

		default:
			conn, err := ln.Accept()
			if err != nil {
				logrus.WithError(err).Error("Failed to accept connection")
			} else {
				go c.HandleConnection(conn)
			}
		}
	}
}

func (c *Connector) HandleConnection(frontendConn net.Conn) {
	defer frontendConn.Close()

	clientAddr := frontendConn.RemoteAddr()

	logrus.
		WithField("client", clientAddr).
		Debug("Got connection")
	defer logrus.WithField("client", clientAddr).Debug("Closing frontend connection")

	// Tee-off the inspected content to a buffer so that we can retransmit it to the backend connection
	inspectionBuffer := new(bytes.Buffer)
	inspectionReader := io.TeeReader(frontendConn, inspectionBuffer)

	bufferedReader := bufio.NewReader(inspectionReader)

	if err := frontendConn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		logrus.
			WithError(err).
			WithField("client", clientAddr).
			Error("Failed to set read deadline")
		return
	}
	packet, err := mcproto.ReadPacket(bufferedReader, clientAddr, c.state)
	if err != nil {
		logrus.WithError(err).WithField("clientAddr", clientAddr).Error("Failed to read packet")
		return
	}

	logrus.
		WithField("client", clientAddr).
		WithField("length", packet.Length).
		WithField("packetID", packet.PacketID).
		Debug("Got packet")

	switch packet.PacketID {
	case mcproto.PacketIdHandshake:
		handshake, err := mcproto.DecodeHandshake(packet.Data)
		if err != nil {
			logrus.WithError(err).WithField("clientAddr", clientAddr).
				Error("Failed to read handshake")
			return
		}

		logrus.
			WithField("client", clientAddr).
			WithField("handshake", handshake).
			Debug("Got handshake")

		var playerInfo *PlayerInfo = nil
		if handshake.NextState == mcproto.StateLogin {
			playerInfo, err = c.readPlayerInfo(handshake.ProtocolVersion, bufferedReader, clientAddr, handshake.NextState)
			if err != nil {
				if errors.Is(err, io.EOF) {
					logrus.
						WithError(err).
						WithField("clientAddr", clientAddr).
						WithField("player", playerInfo).
						Warn("Truncated buffer while reading player info")
				} else {
					logrus.
						WithError(err).
						WithField("clientAddr", clientAddr).
						Error("Failed to read user info")
					return
				}
			}
			logrus.
				WithField("client", clientAddr).
				WithField("player", playerInfo).
				Debug("Got user info")
		}

		c.findAndConnectBackend(frontendConn, clientAddr, inspectionBuffer, handshake.ServerAddress, playerInfo, handshake.NextState)

	case mcproto.PacketIdLegacyServerListPing:
		handshake, ok := packet.Data.(*mcproto.LegacyServerListPing)
		if !ok {
			logrus.
				WithField("client", clientAddr).
				WithField("packet", packet).
				Warn("Unexpected data type for PacketIdLegacyServerListPing")
			return
		}

		logrus.
			WithField("client", clientAddr).
			WithField("handshake", handshake).
			Debug("Got legacy server list ping")

		serverAddress := handshake.ServerAddress

		c.findAndConnectBackend(frontendConn, clientAddr, inspectionBuffer, serverAddress, nil, mcproto.StateStatus)
	default:
		logrus.
			WithField("client", clientAddr).
			WithField("packetID", packet.PacketID).
			Error("Unexpected packetID, expected handshake")
		return
	}
}

func (c *Connector) readPlayerInfo(protocolVersion mcproto.ProtocolVersion, bufferedReader *bufio.Reader, clientAddr net.Addr, state mcproto.State) (*PlayerInfo, error) {
	loginPacket, err := mcproto.ReadPacket(bufferedReader, clientAddr, state)
	if err != nil {
		return nil, fmt.Errorf("failed to read login packet: %w", err)
	}

	if loginPacket.PacketID == mcproto.PacketIdLogin {
		loginStart, err := mcproto.DecodeLoginStart(protocolVersion, loginPacket.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode login start: %w", err)
		}
		return &PlayerInfo{
			Name: loginStart.Name,
			Uuid: loginStart.PlayerUuid,
		}, nil
	} else {
		return nil, fmt.Errorf("expected login packet, got %d", loginPacket.PacketID)
	}
}

func (c *Connector) findAndConnectBackend(frontendConn net.Conn,
	clientAddr net.Addr, preReadContent io.Reader, serverAddress string, playerInfo *PlayerInfo, nextState mcproto.State) {

	logrus.
		WithField("client", clientAddr).
		WithField("server", serverAddress).
		WithField("player", playerInfo).
		WithField("nextState", nextState).
		Info("Handling connection request")

	switch nextState {
	case mcproto.StateStatus:
		c.handleStatusRequest(frontendConn, clientAddr, serverAddress)
	case mcproto.StateLogin:
		c.handleLoginRequest(frontendConn, clientAddr, serverAddress, playerInfo)
	default:
		logrus.
			WithField("client", clientAddr).
			WithField("nextState", nextState).
			Warn("Unexpected next state")
	}
}

func (c *Connector) handleStatusRequest(frontendConn net.Conn, clientAddr net.Addr, serverAddress string) {
	logrus.
		WithField("client", clientAddr).
		WithField("server", serverAddress).
		Info("Handling status request")

	// Clear the read deadline since we'll be doing multiple reads
	if err := frontendConn.SetReadDeadline(noDeadline); err != nil {
		logrus.
			WithError(err).
			WithField("client", clientAddr).
			Error("Failed to clear read deadline")
		return
	}

	bufferedReader := bufio.NewReader(frontendConn)

	statusPacket, err := mcproto.ReadPacket(bufferedReader, clientAddr, mcproto.StateStatus)
	if err != nil {
		logrus.WithError(err).WithField("client", clientAddr).Error("Failed to read status packet")
		return
	}

	if statusPacket.PacketID == mcproto.PacketIdStatusRequest {
		currentMOTD := c.motdManager.GetCurrentMOTD()

		err = mcproto.WriteStatusResponse(frontendConn,
			currentMOTD,
			c.config.ServerStatus.MaxPlayers,
			0,
			c.config.ServerStatus.Version,
			c.config.ServerStatus.Protocol)
		if err != nil {
			logrus.WithError(err).WithField("client", clientAddr).Error("Failed to write status response")
			return
		}

		// Wait for ping request
		pingPacket, err := mcproto.ReadPacket(bufferedReader, clientAddr, mcproto.StateStatus)
		if err != nil {
			logrus.WithError(err).WithField("client", clientAddr).Error("Failed to read ping packet")
			return
		}

		if pingPacket.PacketID == mcproto.PacketIdPingRequest {
			data := pingPacket.Data.([]byte)
			if len(data) >= 8 {
				// Extract the 8-byte timestamp payload
				var payload int64
				for i := 0; i < 8; i++ {
					payload = (payload << 8) | int64(data[i])
				}

				err = mcproto.WritePong(frontendConn, payload)
				if err != nil {
					logrus.WithError(err).WithField("client", clientAddr).Error("Failed to write pong response")
					return
				}
			}
		}

		logrus.
			WithField("client", clientAddr).
			WithField("server", serverAddress).
			WithField("motd", currentMOTD).
			Info("Successfully handled status request")
	}
}

func (c *Connector) handleLoginRequest(frontendConn net.Conn, clientAddr net.Addr, serverAddress string, playerInfo *PlayerInfo) {
	c.motdManager.OnJoinAttempt()

	logrus.
		WithField("client", clientAddr).
		WithField("server", serverAddress).
		WithField("player", playerInfo).
		Info("Handling login request - server is starting up")

	disconnectReason := "🚀 Server is waking up! Please try again in a few minutes."
	err := mcproto.WriteDisconnect(frontendConn, disconnectReason)
	if err != nil {
		logrus.WithError(err).WithField("client", clientAddr).Error("Failed to write disconnect packet")
		return
	}

	logrus.
		WithField("client", clientAddr).
		WithField("server", serverAddress).
		WithField("player", playerInfo).
		WithField("reason", disconnectReason).
		Info("Disconnected player with startup message")
}
