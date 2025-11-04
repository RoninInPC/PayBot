package qrcode

import (
	"errors"
	goqr "github.com/liyue201/goqr"
	"image"
	"main/internal/entity"
	"os"
	"strconv"
	"strings"
)

type QRCode struct {
	qrcodeBytes []byte
	Payment
}

func (qr *QRCode) analyseTextPayment(text string) error {
	var err error
	err = nil
	if !strings.Contains(text, "|") && !strings.HasPrefix(text, "ST000") {
		return errors.New("Is not a QR code requisite text")
	}
	splittedText := strings.Split(text, "|")
	for _, partSplittedText := range splittedText {
		if strings.HasPrefix(partSplittedText, "Name=") {
			qr.Name = strings.Replace(partSplittedText, "Name=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "PersonalAcc=") {
			qr.PersonalAcc = strings.Replace(partSplittedText, "PersonalAcc=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "BankName=") {
			qr.BankName = strings.Replace(partSplittedText, "BankName=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "BIC=") {
			qr.BIC = strings.Replace(partSplittedText, "BIC=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "CorrespAcc=") {
			qr.CorrespAcc = strings.Replace(partSplittedText, "CorrespAcc=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "KPP=") {
			qr.KPP = strings.Replace(partSplittedText, "KPP=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "PayeeINN=") {
			qr.PayeeINN = strings.Replace(partSplittedText, "PayeeINN=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "Purpose=") {
			qr.Purpose = strings.Replace(partSplittedText, "Purpose=", "", 1)
		}
		if strings.HasPrefix(partSplittedText, "Sum=") {
			qr.Sum, err = strconv.ParseFloat(strings.Replace(partSplittedText, "Sum=", "", 1), 10)
		}
	}
	return err
}

func (qr *QRCode) AnalyseFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	return qr.AnalyseImage(img)
}

func (qr *QRCode) AnalyseImage(image image.Image) error {
	qrcodeRecognize, _ := goqr.Recognize(image)
	for _, qrcodeRecognizePart := range qrcodeRecognize {
		return qr.analyseTextPayment(string(qrcodeRecognizePart.Payload))
	}
	return errors.New("Don't be a qrcode")
}

func (qr *QRCode) AnalyseText(text string) error {
	return qr.analyseTextPayment(text)
}

func (qr *QRCode) SaveImage(filename string) error {
	qr.qrcodeBytes, _ = qr.Png(Windows1251, 512)
	return qr.PngFile(filename, Windows1251, 512)
}

func (qr *QRCode) ToEntity(name string) entity.Requisite {
	str, _ := qr.String(Windows1251)
	return entity.Requisite{Name: name, PhotoData: string(qr.qrcodeBytes), Content: str}
}

func (qr *QRCode) FromEntity(req *entity.Requisite) error {
	return qr.analyseTextPayment(req.Content)
}
