package transport

type Transporter interface {
	Transport(request []byte) (response []byte, err error)
}
