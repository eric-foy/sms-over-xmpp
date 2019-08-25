package sms

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/xml"
	"log"
	"strings"

	xco "github.com/mndrix/go-xco"
	errors "github.com/pkg/errors"
)

// gatewayProcess is the piece which sits between the XMPP and HTTP
// processes, translating between their different protocols.
type gatewayProcess struct {
	// fields shared with Component. see docs there
	config *Config
	smsRx  <-chan *Sms
	xmppRx <-chan *xco.Message
	xmppTx chan<- *xco.Message
}

func (g *gatewayProcess) run() <-chan struct{} {
	healthCh := make(chan struct{})
	go g.loop(healthCh)
	return healthCh
}

func (g *gatewayProcess) loop(healthCh chan<- struct{}) {
	defer func() { close(healthCh) }()

	for {
		select {
		case rxSms := <-g.smsRx:
			g.sms2xmpp(rxSms)
		case msg := <-g.xmppRx:
			err := g.xmpp2sms(msg)
			if err != nil {
				log.Printf("ERROR: converting XMPP to SMS: %s", err)
				return
			}
		}
	}
}

func (g *gatewayProcess) sms2xmpp(sms *Sms) {
	to, err := xco.ParseAddress(g.config.Xmpp.JID)
	if err != nil {
		log.Printf("ERROR: parsing JID: %s", err)
	}

	msg := &xco.Message{
		XMLName: xml.Name{
			Local: "message",
			Space: "jabber:component:accept",
		},
		Header: xco.Header{
			ID: NewId(),
			To: to,
			From: xco.Address{
				LocalPart:  sms.From,
				DomainPart: g.config.Xmpp.Domain,
			},
		},
		Type: "chat",
		Body: sms.Body,
	}

	go func() { g.xmppTx <- msg }()
}

func (g *gatewayProcess) xmpp2sms(m *xco.Message) error {
	var err error
	sms := &Sms{
		Body: m.Body,
		To:   m.To.LocalPart,
	}

	// choose an SMS provider
	provider, err := g.config.SmsProvider()
	if err != nil {
		return errors.Wrap(err, "choosing an SMS provider")
	}

	// send the message
	id, err := provider.SendSms(sms)
	if err != nil {
		return errors.Wrap(err, "sending SMS")
	}
	log.Printf("Sent SMS with ID %s", id)

	return nil
}

// NewId generates a random string which is suitable as an XMPP stanza
// ID.  The string contains enough entropy to be universally unique.
func NewId() string {
	// generate 128 random bits (6 more than standard UUID)
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	// convert them to base 32 encoding
	s := base32.StdEncoding.EncodeToString(bytes)
	return strings.ToLower(strings.TrimRight(s, "="))
}
