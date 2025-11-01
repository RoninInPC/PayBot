package qrcode

import "image"

type QRCode struct {
}

func (qr *QRCode) AnalyseLink(link string) error {
	panic("Not Implemented")
}

func (qr *QRCode) AnalyseImage(img image.Image) error {
	panic("Not Implemented")
}

func (qr *QRCode) AnalyseText(text string) error {
	panic("Not Implemented")
}

func (qr *QRCode) SaveImage(filename string) error {
	panic("Not Implemented")
}
