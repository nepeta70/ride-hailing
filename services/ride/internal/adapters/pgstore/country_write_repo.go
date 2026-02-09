package pgstore

type CountryWriteEntity struct {
	CountryReadEntity
	IsEnabled bool `db:"is_enabled" json:"is_enabled"`
}
