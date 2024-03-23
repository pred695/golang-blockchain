package Blockchain

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	ID        []byte //references to the previous output that led to the input
	OutputIdx int    //index of the referenced output which is spent in the transaction
	Sign      string //signature or the sender's public key
}

func (inputTx *TxInput) CanUnlock(data string) bool {
	return inputTx.Sign == data
}

func (outputTx *TxOutput) CorrectRecAddress(data string) bool {
	return outputTx.PubKey == data
}
