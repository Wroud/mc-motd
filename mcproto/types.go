package mcproto

import (
	"fmt"

	"github.com/google/uuid"
)

type Frame struct {
	Length  int
	Payload []byte
}

type State int

/*
Handshaking -> Status
Handshaking -> Login -> ...
*/
const (
	StateHandshaking State = 0
	StateStatus      State = 1
	StateLogin       State = 2
)

var trimLimit = 64

func trimBytes(data []byte) ([]byte, string) {
	if len(data) < trimLimit {
		return data, ""
	} else {
		return data[:trimLimit], "..."
	}
}

func (f *Frame) String() string {
	trimmed, cont := trimBytes(f.Payload)
	return fmt.Sprintf("Frame:[len=%d, payload=%#X%s]", f.Length, trimmed, cont)
}

type Packet struct {
	Length   int
	PacketID int
	// Data is either a byte slice of raw content or a decoded message
	Data interface{}
}

func (p *Packet) String() string {
	if dataBytes, ok := p.Data.([]byte); ok {
		trimmed, cont := trimBytes(dataBytes)
		return fmt.Sprintf("Frame:[len=%d, packetId=%d, data=%#X%s]", p.Length, p.PacketID, trimmed, cont)
	} else {
		return fmt.Sprintf("Frame:[len=%d, packetId=%d, data=%+v]", p.Length, p.PacketID, p.Data)
	}
}

type ProtocolVersion int

// Source: https://minecraft.wiki/w/Minecraft_Wiki:Projects/wiki.vg_merge/Protocol_History
const (
	// ProtocolVersion1_18_2 is the protocol version for Minecraft 1.18.2
	// Docs: https://minecraft.wiki/w/Java_Edition_protocol/Packets?oldid=2772791
	ProtocolVersion1_18_2 ProtocolVersion = 758
	// ProtocolVersion1_19 is the protocol version for Minecraft 1.19
	// Docs: https://minecraft.wiki/w/Java_Edition_protocol/Packets?oldid=2772904
	ProtocolVersion1_19 ProtocolVersion = 759
	// ProtocolVersion1_19_2 is the protocol version for Minecraft 1.19.2
	// Docs: https://minecraft.wiki/w/Java_Edition_protocol/Packets?oldid=2772944
	ProtocolVersion1_19_2 ProtocolVersion = 760
	// ProtocolVersion1_19_3 is the protocol version for Minecraft 1.19.3
	ProtocolVersion1_19_3 ProtocolVersion = 761
	// ProtocolVersion1_19_4 is the protocol version for Minecraft 1.19.4
	ProtocolVersion1_19_4 ProtocolVersion = 762
	// ProtocolVersion1_20 is the protocol version for Minecraft 1.20/1.20.1
	ProtocolVersion1_20 ProtocolVersion = 763
	// ProtocolVersion1_20_2 is the protocol version for Minecraft 1.20.2
	ProtocolVersion1_20_2 ProtocolVersion = 764
	// ProtocolVersion1_20_3 is the protocol version for Minecraft 1.20.3/1.20.4
	ProtocolVersion1_20_3 ProtocolVersion = 765
	// ProtocolVersion1_20_5 is the protocol version for Minecraft 1.20.5
	ProtocolVersion1_20_5 ProtocolVersion = 766
	// ProtocolVersion1_21 is the protocol version for Minecraft 1.21/1.21.1
	ProtocolVersion1_21 ProtocolVersion = 767
	// ProtocolVersion1_21_2 is the protocol version for Minecraft 1.21.2/1.21.3
	ProtocolVersion1_21_2 ProtocolVersion = 768
	// ProtocolVersion1_21_4 is the protocol version for Minecraft 1.21.4
	ProtocolVersion1_21_4 ProtocolVersion = 769
	// ProtocolVersion1_21_5 is the protocol version for Minecraft 1.21.5
	ProtocolVersion1_21_5 ProtocolVersion = 770
	// ProtocolVersion1_21_6 is the protocol version for Minecraft 1.21.6
	ProtocolVersion1_21_6 ProtocolVersion = 771
	// ProtocolVersion1_21_7 is the protocol version for Minecraft 1.21.7/1.21.8
	ProtocolVersion1_21_7 ProtocolVersion = 772
)

const (
	PacketIdHandshake            = 0x00
	PacketIdLogin                = 0x00 // during StateLogin
	PacketIdLegacyServerListPing = 0xFE
	// Status state packets
	PacketIdStatusRequest = 0x00 // during StateStatus
	PacketIdPingRequest   = 0x01 // during StateStatus
)

type Handshake struct {
	ProtocolVersion ProtocolVersion
	ServerAddress   string
	ServerPort      uint16
	NextState       State
}

type LoginStart struct {
	Name       string
	PlayerUuid uuid.UUID
}

func NewLoginStart() *LoginStart {
	return &LoginStart{
		// Note: This is indistinguishable between no UUID provided, and a provided UUID of all 0s
		PlayerUuid: uuid.Nil,
	}
}

type LegacyServerListPing struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
}

type ByteReader interface {
	ReadByte() (byte, error)
}

const (
	PacketLengthFieldBytes = 1
)

// VersionToProtocol maps Minecraft version strings to their corresponding protocol numbers
func VersionToProtocol(version string) (int, bool) {
	versionMap := map[string]int{
		"1.18.2": int(ProtocolVersion1_18_2), // 758
		"1.19":   int(ProtocolVersion1_19),   // 759
		"1.19.0": int(ProtocolVersion1_19),   // 759
		"1.19.1": int(ProtocolVersion1_19_2), // 760 (1.19.1 uses same as 1.19.2)
		"1.19.2": int(ProtocolVersion1_19_2), // 760
		"1.19.3": int(ProtocolVersion1_19_3), // 761
		"1.19.4": int(ProtocolVersion1_19_4), // 762
		"1.20":   int(ProtocolVersion1_20),   // 763
		"1.20.0": int(ProtocolVersion1_20),   // 763
		"1.20.1": int(ProtocolVersion1_20),   // 763
		"1.20.2": int(ProtocolVersion1_20_2), // 764
		"1.20.3": int(ProtocolVersion1_20_3), // 765
		"1.20.4": int(ProtocolVersion1_20_3), // 765 (1.20.4 uses same as 1.20.3)
		"1.20.5": int(ProtocolVersion1_20_5), // 766
		"1.20.6": int(ProtocolVersion1_20_5), // 766 (1.20.6 uses same as 1.20.5)
		"1.21":   int(ProtocolVersion1_21),   // 767
		"1.21.0": int(ProtocolVersion1_21),   // 767
		"1.21.1": int(ProtocolVersion1_21),   // 767
		"1.21.2": int(ProtocolVersion1_21_2), // 768
		"1.21.3": int(ProtocolVersion1_21_2), // 768 (1.21.3 uses same as 1.21.2)
		"1.21.4": int(ProtocolVersion1_21_4), // 769
		"1.21.5": int(ProtocolVersion1_21_5), // 770
		"1.21.6": int(ProtocolVersion1_21_6), // 771
		"1.21.7": int(ProtocolVersion1_21_7), // 772
		"1.21.8": int(ProtocolVersion1_21_7), // 772 (1.21.8 uses same as 1.21.7)
	}

	protocol, exists := versionMap[version]
	return protocol, exists
}
