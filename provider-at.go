package sms

import (
	"github.com/eric-foy/go-gsm-lib"
)

type AT struct {
	phoneNum string
	modem    *gsm.Modem
}

func (at *AT) RunPstnProcess(rxSmsCh chan<- RxSms) <-chan struct{} {
	healthCh := make(chan struct{})
	go func() {
		defer func() { close(healthCh) }()
		for {
			cmt := <-at.modem.Cmt

			sms := &Sms{
				From: cmt.Oa,
				To:   at.phoneNum,
				Body: cmt.Data,
			}

			rx := &rxSmsMessage{
				sms: sms,
			}

			rxSmsCh <- rx
		}
	}()

	return healthCh
}

func (at *AT) SendSms(sms *Sms) (string, error) {
	return at.modem.SendSMS(sms.To, sms.Body).Mr, nil
}
