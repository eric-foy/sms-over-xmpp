package sms

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// Kannel represents an account with the communications provider Kannel.
type Kannel struct {
	// where to send outgoing HTTP requests
	SendsmsHost string
	SendsmsPort string

	// credentials for HTTP auth
	httpUsername string
	httpPassword string

	client *http.Client
}

// make sure we implement the right interfaces
var _ SmsProvider = &Kannel{}

func (t *Kannel) httpClient() *http.Client {
	if t.client == nil {
		return http.DefaultClient
	}
	return t.client
}

// do runs a single API request against Kannel
func (t *Kannel) do(service string, form url.Values) (string, error) {
	//url := "https://" + t.SendsmsHost + ":" + t.SendsmsPort + "/" + service + "?" + form.Encode()
	url := "http://localhost:13013/cgi-bin/" + service + "?" + form.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "building HTTP request")
	}

	res, err := t.httpClient().Do(req)
	if err != nil {
		return "", errors.Wrap(err, "running HTTP request")
	}

	defer res.Body.Close()

	// parse response

	// was the message queued for delivery?

	id := xid.New()

	return id.String(), nil
}

func (t *Kannel) SendSms(sms *Sms) (string, error) {
	form := make(url.Values)
	form.Set("username", t.httpUsername)
	form.Set("password", t.httpPassword)
	form.Set("to", sms.To)
	//form.Set("from", sms.From)
	form.Set("text", sms.Body)
	res, err := t.do("sendsms", form)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (t *Kannel) RunPstnProcess(rxSmsCh chan<- RxSms) <-chan struct{} {
	healthCh := make(chan struct{})
	return healthCh
}
