package enums

type SenderType string

const (
	SenderTypeDriver    SenderType = "driver"
	SenderTypeRider     SenderType = "rider"
	SenderTypeUser      SenderType = "user" // rider or driver
	SenderTypeAdmin     SenderType = "admin"
	SenderTypeAnonymous SenderType = "anonymous"
	SenderTypeService   SenderType = "service"
)

func (u SenderType) IsValid() bool {
	switch u {
	case SenderTypeDriver, SenderTypeRider, SenderTypeUser, SenderTypeAdmin, SenderTypeAnonymous, SenderTypeService:
		return true
	default:
		return false
	}
}

func (u SenderType) String() string {
	return string(u)
}
