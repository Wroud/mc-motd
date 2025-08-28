package server

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
	Protocol        int    `default:"772" usage:"The protocol version number (772 for 1.21.8, 770 for 1.21.5, etc.)"`
}

type Config struct {
	Port         int                `default:"25565" usage:"The [port] bound to listen for Minecraft client connections"`
	Webhook      WebhookConfig      `usage:"Webhook configuration"`
	ServerStatus ServerStatusConfig `usage:"Server status configuration for server list responses"`
}
