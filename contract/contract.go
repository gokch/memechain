package contract

import "github.com/gokch/memechain/types"

type Contract interface {
	Fee() uint64
	Execute() error
}

var Contracts = map[types.AccountID]Contract{
	{0x01}: &New{},
}

type New struct {
}

func (c *New) Fee() uint64 {
	return 0
}

func (c *New) Execute() error {
	return nil
}
