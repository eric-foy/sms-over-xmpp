package sms

import (
	"regexp"

	gsm "github.com/eric-foy/go-gsm-lib"
	xco "github.com/mndrix/go-xco"
	errors "github.com/pkg/errors"
)

// StaticConfig intends to implement the Config interface
var _ Config = &StaticConfig{}

type StaticConfig struct {
	Http HttpConfig `toml:"http"`

	Xmpp StaticConfigXmpp `toml:"xmpp"`

	// Phones maps an E.164 phone number to an XMPP address.  If a
	// mapping is not found here, the inverse of Users is considered.
	Phones map[string]string `toml:"phones"`

	// Users maps an XMPP address to an E.164 phone number.
	Users map[string]string `toml:"users"`

	// AT contains optional configuration for connecting with modem via AT commands.
	AT *ATConfig `toml:"at"`
	// Can only connect open one connection to modem at a time, save address.
	// Used for both sending SMS and receiving.
	SlottedAT *AT
}

type HttpConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`

	Username string `toml:"username"`
	Password string `toml:"password"`

	PublicUrl string `toml:"public-url"`
}

type StaticConfigXmpp struct {
	Host   string `toml:"host"`
	Name   string `toml:"name"`
	Port   int    `toml:"port"`
	Secret string `toml:"secret"`
}

type ATConfig struct {
	// Available options: serial_tcp, serial
	Method string `toml:"method"`

	// Such as /dev/ttyAMA0 for serial and 192.168.1.111:7777 for serial_tcp.
	Device string `toml:"device"`

	// TODO phone number could be grabbed from Users section of config.
	PhoneNum string `toml:"my-number"`
}

func (self *StaticConfig) ComponentName() string {
	return self.Xmpp.Name
}

func (self *StaticConfig) SharedSecret() string {
	return self.Xmpp.Secret
}

func (self *StaticConfig) HttpHost() string {
	host := self.Http.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return host
}

func (self *StaticConfig) HttpPort() int {
	port := self.Http.Port
	if port == 0 {
		port = 9677
	}
	return port
}

func (self *StaticConfig) XmppHost() string {
	return self.Xmpp.Host
}

func (self *StaticConfig) XmppPort() int {
	return self.Xmpp.Port
}

func (self *StaticConfig) AddressToPhone(addr xco.Address) (string, error) {
	e164, ok := self.Users[addr.LocalPart+"@"+addr.DomainPart]
	if ok {
		return e164, nil
	}

	// maybe the XMPP local part is a phone number
	matched, err := regexp.MatchString(`[0-9]{9}`, addr.LocalPart)
	if err != nil {
		return "", err
	}
	if matched {
		return addr.LocalPart, nil
	}

	return "", ErrIgnoreMessage
}

func (self *StaticConfig) PhoneToAddress(e164 string) (xco.Address, error) {
	// is there an explicit mapping?
	jid, ok := self.Phones[e164]
	if ok {
		return xco.ParseAddress(jid)
	}

	// maybe there's an implicit mapping
	for jid, phone := range self.Users {
		if phone == e164 {
			return xco.ParseAddress(jid)
		}
	}

	return xco.Address{}, ErrIgnoreMessage
}

func (self *StaticConfig) SmsProvider() (SmsProvider, error) {
	if self.AT == nil {
		return nil, errors.New("Need to configure an SMS provider")
	}

	if self.SlottedAT == nil {
		modem, err := gsm.New(self.AT.Method, self.AT.Device)
		if err != nil {
			return nil, errors.Wrap(err, "Trouble connecting with AT device")
		}
		self.SlottedAT = &AT{
			phoneNum: self.AT.PhoneNum,
			modem:    modem,
		}
	}

	return self.SlottedAT, nil
}
