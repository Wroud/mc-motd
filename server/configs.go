package server

import "github.com/wroud/mc-motd/mcproto"

type WebhookConfig struct {
	Url         string `usage:"If set, a POST request that contains connection status notifications will be sent to this HTTP address"`
	RequireUser bool   `default:"false" usage:"Indicates if the webhook will only be called if a user is connecting rather than just server list/ping"`
}

type ServerStatusConfig struct {
	SleepingMOTD    string `default:"🌙 Server sleeping, join to wake up!" usage:"The MOTD displayed when the server is in sleeping state"`
	StartingMOTD    string `default:"⚡ Server starting up..." usage:"The MOTD displayed when the server is starting up"`
	StartingTimeout int    `default:"300" usage:"How many seconds to show the starting MOTD after a join attempt (default: 5 minutes)"`
	MaxPlayers      int    `default:"20" usage:"The maximum number of players displayed in the server list"`
	Version         string `default:"1.21.8" usage:"The Minecraft version displayed in the server list"`
	Protocol        int    `default:"0" usage:"The protocol version number. If 0 (default), will be auto-detected from Version. Set explicitly to override (e.g., 772 for 1.21.8, 770 for 1.21.5)"`
}

// GetProtocol returns the protocol version to use.
// If Protocol is explicitly set (non-zero), it uses that value.
// Otherwise, it attempts to detect the protocol from the Version string.
// Falls back to protocol 772 (1.21.8) if version detection fails.
func (c *ServerStatusConfig) GetProtocol() int {
	// If Protocol is explicitly set (non-zero), use it
	if c.Protocol != 0 {
		return c.Protocol
	}

	// Try to detect protocol from version string
	if detectedProtocol, found := mcproto.VersionToProtocol(c.Version); found {
		return detectedProtocol
	}

	// Fallback to default protocol for 1.21.8 if detection fails
	return 772
}

type Config struct {
	Port         int                `default:"25565" usage:"The [port] bound to listen for Minecraft client connections"`
	Webhook      WebhookConfig      `usage:"Webhook configuration"`
	ServerStatus ServerStatusConfig `usage:"Server status configuration for server list responses"`
}
