// Package sms provides an XMPP component (XEP-0114) which acts as a
// gateway or proxy between XMPP and SMS.  It allows you to send and
// receive SMS messages as if they were XMPP messages.  This lets you
// interact with the SMS network using your favorite XMPP client.
//
// Many users will be satisfied to run the sms-over-xmpp command with
// an appropriate configuration file.  This Go package is intended for
// those who want greater control over their SMS gateway or who wish
// to incorporate the XMPP component into existing Go code.
package sms // import "github.com/eric-foy/sms-over-xmpp"

import (
	"fmt"
	"log"
	"time"

	xco "github.com/mndrix/go-xco"
	errors "github.com/pkg/errors"
)

// ErrIgnoreMessage should be returned to indicate that a message
// should be ignored; as if it never happened.
var ErrIgnoreMessage = errors.New("ignore this message")

// Component represents an SMS-over-XMPP component.
type Component struct {
	config Config

	// receiptFor contains message delivery receipts that
	// haven't been delivered yet.  the key is a provider's outgoing
	// SMS identifier.  the value is the delivery receipt that we should deliver
	// once the associated SMS has been delivered.
	receiptFor map[string]*xco.Message

	// rxSmsCh is a channel connecting PSTN->Gateway.  It communicates
	// information received about SMS (a message, a status update,
	// etc.)
	rxSmsCh chan RxSms

	// rxXmppCh is a channel connecting XMPP->Gateway. It communicates
	// incoming XMPP messages.  It doesn't carry other XMPP stanzas
	// (Iq, Presence, etc) since those are handled inside the XMPP
	// process.
	rxXmppCh chan *xco.Message

	// txXmppCh is a channel connecting Gateway->XMPP. It communicates
	// outgoing XMPP messages.
	txXmppCh chan *xco.Message
}

// Main runs a component using the given configuration.  It's the main
// entrypoint for launching your own component if you don't want to
// use the sms-over-xmpp command.
func Main(config Config) {
	sc := &Component{config: config}
	sc.receiptFor = make(map[string]*xco.Message)
	sc.rxSmsCh = make(chan RxSms)
	sc.rxXmppCh = make(chan *xco.Message)
	sc.txXmppCh = make(chan *xco.Message)

	// start processes running
	gatewayDead := sc.runGatewayProcess()
	xmppDead := sc.runXmppProcess()
	pstnDead := sc.runPstnProcess()

	for {
		select {
		case _ = <-gatewayDead:
			log.Printf("Gateway died. Restarting")
			gatewayDead = sc.runGatewayProcess()
		case _ = <-pstnDead:
			log.Printf("PSTN died. Restarting")
			pstnDead = sc.runPstnProcess()
		case _ = <-xmppDead:
			log.Printf("XMPP died. Restarting")
			time.Sleep(1 * time.Second) // don't hammer server
			xmppDead = sc.runXmppProcess()
		}
	}
}

// runGatewayProcess starts the Gateway process. it translates between
// the PSTN and XMPP processes.
func (sc *Component) runGatewayProcess() <-chan struct{} {
	gateway := &gatewayProcess{
		// as long as it's alive, Gateway owns these values
		config:     sc.config,
		receiptFor: sc.receiptFor,
		smsRx:      sc.rxSmsCh,
		xmppRx:     sc.rxXmppCh,
		xmppTx:     sc.txXmppCh,
	}
	return gateway.run()
}

// runPstnProcess starts the PSTN process
func (sc *Component) runPstnProcess() <-chan struct{} {
	config := sc.config

	// choose an SMS provider
	provider, err := config.SmsProvider()
	if err != nil {
		msg := fmt.Sprintf("Couldn't choose an SMS provider: %s", err)
		panic(msg)
	}

	return provider.RunPstnProcess(sc.rxSmsCh)
}

// runXmppProcess starts the XMPP process
func (sc *Component) runXmppProcess() <-chan struct{} {
	x := &xmppProcess{
		host:   sc.config.XmppHost(),
		port:   sc.config.XmppPort(),
		name:   sc.config.ComponentName(),
		secret: sc.config.SharedSecret(),

		gatewayTx: sc.txXmppCh,
		gatewayRx: sc.rxXmppCh,
	}
	return x.run()
}
