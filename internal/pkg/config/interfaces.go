package config

type Validator interface {
	Validate() error
}

type Initializer interface {
	Init() error
}
