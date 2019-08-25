package sms

type Config struct {
	Xmpp ConfigXmpp `toml:"xmpp"`

	// AT contains optional configuration for connecting with modem via AT commands.
	AT *ATConfig `toml:"at"`
}

type ConfigXmpp struct {
	Host   string `toml:"host"`
	Domain string `toml:"domain"`
	JID    string `toml:"jid"`
	Port   int    `toml:"port"`
	Secret string `toml:"secret"`
}

type ATConfig struct {
	// Available options: serial_tcp, serial
	Method string `toml:"method"`

	// Such as /dev/ttyAMA0 for serial and 192.168.1.111:7777 for serial_tcp.
	Device string `toml:"device"`
}
