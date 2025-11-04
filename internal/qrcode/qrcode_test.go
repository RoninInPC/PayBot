package qrcode

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	qrcode := QRCode{}
	err := qrcode.AnalyseFile("qr.png")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(qrcode.Payment)
	}
	qrcode.Sum = 4800.23
	fmt.Println(qrcode.Payment)
	fmt.Println(qrcode.String(1))
	qrcode.SaveImage("qr2.png")
}

func Test2(t *testing.T) {

}
