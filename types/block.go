package types

type Block struct {
	Hash   []byte
	Header *BlockHeader
	Body   *BlockBody
}

type BlockHeader struct {
	Height        int
	Timestamp     int64
	PrevBlockHash []byte
	Sign          []byte
	DirRoot       []byte
	CoinBase      []byte
}

type BlockBody struct {
	Transactions []*Transaction
}

// TODO : Global directory which contain all subject and file
// root hash contain in block header
type Directory struct {
	Name           string
	Hash           HashID
	SubDirectories []*Directory
}
