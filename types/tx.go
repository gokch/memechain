package types

type TxType int32

const (
	TxTypeTransfer TxType = iota
	TxTypeFileDeploy
	TxTypeFileSend
	TxTypeFileReSend
)

type Transaction struct {
	From []byte
	To   []byte
	Sign []byte
	Fee  int64
}
