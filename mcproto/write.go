package mcproto

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// WriteVarInt writes a variable-length integer to the writer
func WriteVarInt(writer io.Writer, value int) error {
	for {
		temp := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			temp |= 0x80
		}
		if err := binary.Write(writer, binary.BigEndian, temp); err != nil {
			return err
		}
		if value == 0 {
			break
		}
	}
	return nil
}

// WriteString writes a string as VarInt length + UTF-8 bytes
func WriteString(writer io.Writer, s string) error {
	data := []byte(s)
	if err := WriteVarInt(writer, len(data)); err != nil {
		return err
	}
	_, err := writer.Write(data)
	return err
}

// WritePacket writes a packet with packet ID and data
func WritePacket(writer io.Writer, packetID int, data []byte) error {
	buf := new(bytes.Buffer)

	// Write packet ID
	if err := WriteVarInt(buf, packetID); err != nil {
		return err
	}

	// Write data
	if _, err := buf.Write(data); err != nil {
		return err
	}

	packetData := buf.Bytes()

	// Write packet length
	if err := WriteVarInt(writer, len(packetData)); err != nil {
		return err
	}

	// Write packet data
	_, err := writer.Write(packetData)
	return err
}

// StatusResponse represents the server status response JSON
type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
}

// WriteStatusResponse writes a status response packet
func WriteStatusResponse(writer io.Writer, motd string, maxPlayers, onlinePlayers int, version string, protocol int) error {
	response := StatusResponse{}
	response.Version.Name = version
	response.Version.Protocol = protocol
	response.Players.Max = maxPlayers
	response.Players.Online = onlinePlayers
	response.Description.Text = motd

	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// Packet ID for Status Response is 0x00 in status state
	buf := new(bytes.Buffer)
	if err := WriteString(buf, string(jsonData)); err != nil {
		return err
	}

	return WritePacket(writer, 0x00, buf.Bytes())
}

// WritePong writes a pong response packet
func WritePong(writer io.Writer, payload int64) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, payload); err != nil {
		return err
	}

	// Packet ID for Pong is 0x01 in status state
	return WritePacket(writer, 0x01, buf.Bytes())
}

// WriteDisconnect writes a disconnect packet during login state
func WriteDisconnect(writer io.Writer, reason string) error {
	// Create JSON chat component for disconnect reason
	reasonJSON := fmt.Sprintf(`{"text":"%s"}`, reason)

	buf := new(bytes.Buffer)
	if err := WriteString(buf, reasonJSON); err != nil {
		return err
	}

	// Packet ID for Disconnect (login) is 0x00 in login state
	return WritePacket(writer, 0x00, buf.Bytes())
}
