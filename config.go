package sms

import (
	gsm "github.com/eric-foy/go-gsm-lib"
	errors "github.com/pkg/errors"
)

type Config struct {
	Xmpp ConfigXmpp `toml:"xmpp"`

	// AT contains optional configuration for connecting with modem via AT commands.
	AT *ATConfig `toml:"at"`
	// Can only connect open one connection to modem at a time, save address.
	// Used for both sending SMS and receiving.
	SlottedAT *AT
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

func (self *Config) SmsProvider() (SmsProvider, error) {
	if self.AT == nil {
		return nil, errors.New("Need to configure an SMS provider")
	}

	if self.SlottedAT == nil {
		modem, err := gsm.New(self.AT.Method, self.AT.Device)
		if err != nil {
			return nil, errors.Wrap(err, "Trouble connecting with AT device")
		}
		self.SlottedAT = &AT{
			modem: modem,
		}
	}

	return self.SlottedAT, nil
}
