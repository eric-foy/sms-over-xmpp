# SMS over XMPP

sms-over-xmpp is an XMPP component (XEP-0114) that acts as a gateway
between an XMPP network and the SMS network.  It allows you to send
and receive SMS messages as if they were XMPP messages, using your
favorite XMPP client.

You send an XMPP message and your friend receives an SMS.  When she
responds by SMS, you receive an XMPP message.

# Prerequisites

You'll need the following available to install and run sms-over-xmpp.

  * [Go](https://golang.org/dl/)
  * An XMPP server (I like [Prosody](http://prosody.im/))
  * A gsm module or a phone that supports AT commands through a serial interface.
  * An activated sim card.

## AT
Using library [go-gsm-lib](https://github.com/eric-foy/go-gsm-lib)

I'm using a SIM800 module on a raspberry pi that has the gsm modem exposed on the /dev/ttyAMA0 tty.
This connects with the help of go-gsm-lib to sms-over-xmpp which in turn is a component of a XMPP server.

My XMPP server is local on the raspberry pi.

(Might expand documentation if bored...)

## XMPP server

sms-over-xmpp is an XMPP component as defined
in [XEP-0114](http://xmpp.org/extensions/xep-0114.html).  That means
that it needs an existing XMPP server to interact with the XMPP
network.  There are several open source XMPP servers available.  My
favorite is [Prosody](http://prosody.im/).  It's easy to configure and
operate.  [ejabberd](https://www.ejabberd.im/) is another popular
choice.

Once your XMPP server is running, you'll need to add an external
component for sms-over-xmpp.  Instructions are available:

  * for [Prosody](http://prosody.im/doc/components#adding_an_external_component)
  * for [ejabberd](https://www.ejabberd.im/node/5134)

You'll need to enter the component's host name and password in your
sms-over-xmpp configuration file later.

# Installation

Install the binary with

    go get github.com/AGWA/sms-over-xmpp/...

Write a config file (`config.toml` is a common name):

```toml
# define how to connect to your XMPP server
[xmpp]
host = "127.0.0.1"
port = 5347
jid = "john@example.com"
name = "sms.example.com"
secret = "shared secret from your XMPP server config"

# AT (not needed if using Twillio)
[at]
method = "serial"
device = "/dev/ttyAMA0"
```

Run your SMS component:

    sms-over-xmpp config.toml
