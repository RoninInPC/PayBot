package entity

type TypeFile int

const (
	Photo      TypeFile = 0
	Doc        TypeFile = 1
	Video      TypeFile = 2
	Audio      TypeFile = 3
	Voice      TypeFile = 4
	VideoVoice TypeFile = 5
	Animation  TypeFile = 6
)

type File struct {
	Filename string
	Type     TypeFile
}

type MessageFromAdminBot struct {
	TelegramID int64
	Text       string
	Files      []File
}

type MessageFromUserBot struct {
	RequisiteContent []byte
	IsFile           bool
	IsImage          bool
	TariffPicked     Tariff
	PromoCodePicked  PromoCode
}
