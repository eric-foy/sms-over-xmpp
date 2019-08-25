package sms

// Sms represents a single SMS message.
type Sms struct {
	// To is the E.164 phone number to which the message was/will-be
	// sent.
	To string

	// From is the E.164 phone number that send/is-sending the
	// message.
	From string

	// Body is the text content of the message.
	Body string
}

// RxSms represents information we've received about an SMS. it could
// be a new message arriving or a status update about a message we
// sent.
type RxSms interface {
	// IsRxSms is a dummy method for tagging those types which
	// represent incoming SMS data.
	IsRxSms()

	// ErrCh returns a channel on which to report errors that happen
	// while processing this SMS.
	ErrCh() chan<- error
}

// rxSmsMessage represents a newly arrived message
type rxSmsMessage struct {
	// sms is the message content
	sms *Sms

	// errCh is the channel for implement ErrCh() method
	errCh chan<- error
}

func MakeRxSmsMessage(sms *Sms, errCh chan<- error) RxSms {
	return &rxSmsMessage{
		sms:   sms,
		errCh: errCh,
	}
}

// implement RxSms interface
var _ RxSms = &rxSmsMessage{}

func (*rxSmsMessage) IsRxSms()              {}
func (i *rxSmsMessage) ErrCh() chan<- error { return i.errCh }
