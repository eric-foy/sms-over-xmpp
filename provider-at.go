package sms

import gsm "github.com/eric-foy/go-gsm-lib"

type AT struct {
	phoneNum string
	modem    *gsm.Modem
}

func (at *AT) RunPstnProcess(rxSmsCh chan<- *Sms) <-chan struct{} {
	go at.modem.ReadTTY()
	go at.modem.InitDevice()
	healthCh := make(chan struct{})
	go func() {
		defer func() { close(healthCh) }()
		for {
			cmt := <-at.modem.Cmt

			rxSmsCh <- &Sms{
				From: cmt.Oa,
				Body: cmt.Data,
			}
		}
	}()

	return healthCh
}

func (at *AT) SendSms(sms *Sms) (string, error) {
	cmgs, err := at.modem.SendSMS(sms.To, sms.Body)
	if err != nil {
		return "", err
	}
	return cmgs.Mr, nil
}
