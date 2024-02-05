package state

type Meme struct {
	MID       string
	Category  []Category
	Providers []Provider
}

type Category struct {
}

type Provider interface {
	Type() string
	GetUrl() (string, error)
	GetData() ([]byte, error)
	GetFunc() (func() ([]byte, error), error)
}
