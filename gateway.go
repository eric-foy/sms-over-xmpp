package sms

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/xml"
	"log"
	"strings"

	gsm "github.com/eric-foy/go-gsm-lib"
	xco "github.com/mndrix/go-xco"
)

// gatewayProcess is the piece which sits between the XMPP and HTTP
// processes, translating between their different protocols.
type gatewayProcess struct {
	// fields shared with Component. see docs there
	config *Config
	modem  *gsm.Modem
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
		case rxAT := <-g.modem.RxAT:
			switch rx := rxAT.(type) {
			case gsm.RxCMT:
				g.sms2xmpp(rx)
			case gsm.RxCMGS:
				log.Printf("Sent SMS with ID %s\n", rx.Mr)
			}
		case msg := <-g.xmppRx:
			g.xmpp2sms(msg)
		}
	}
}

func (g *gatewayProcess) sms2xmpp(sms gsm.RxCMT) {
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
				LocalPart:  sms.Oa,
				DomainPart: g.config.Xmpp.Domain,
			},
		},
		Type: "chat",
		Body: sms.Data,
	}

	go func() { g.xmppTx <- msg }()
}

func (g *gatewayProcess) xmpp2sms(m *xco.Message) {
	cmgs := gsm.TxCMGS{
		Da:   m.To.LocalPart,
		Toda: 145,
		Text: m.Body,
	}
	go func() { g.modem.TxAT <- cmgs }()
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
