package p5

import (
	"encoding/json"
)

type Transaction struct {
	Id              int32           `json:"id"`
	TransactionData TransactionData `json:"transactionData"`
	Hops            int32           `json:"hops"`
}

type TransactionData struct {
	Msg       Msg    `json:"msg"`
	MinerId   int32  `json:"minerId"`
	Signature string `json:"signature"`
}

type Msg struct {
	PayerId int32 `json:"payerId"`
	PayeeId int32 `json:"payeeId"`
	Amount  int32 `json:"amount"`
	TxFee   int32 `json:"txFee"`
	Total   int32 `json:"total"`
}

type TransactionPost struct {
	PayeeId int32 `json:"payeeId"`
	Amount  int32 `json:"amount"`
	TxFee   int32 `json:"txFee"`
}

/**
Create NewTransactionData instance
 */
func NewTransaction(id int32, txData TransactionData, hops int32) Transaction {
	transaction := Transaction{id, txData, hops}
	return transaction
}

/**
Create NewTransactionData instance
 */
func NewMsg(payerId int32, payeeId int32, amount int32, txFee int32, total int32) Msg {
	msg := Msg{payerId, payeeId, amount, txFee, total}
	return msg
}

/**
Encode TransactionData to json
 */
func (msg *Msg) EncodeToJson() (string, error) {
	jsonMsg, err := json.Marshal(msg)
	return string(jsonMsg), err
}

/**
Create NewTransactionData instance
 */
func NewTransactionData(msg Msg, minerId int32, signature string) TransactionData {
	transactionData := TransactionData{msg, minerId, signature}
	return transactionData
}

/**
Encode TransactionPost to json
 */
func (txPost *TransactionPost) EncodeToJson() (string, error) {
	jsonTransactionPost, err := json.Marshal(txPost)
	return string(jsonTransactionPost), err
}

/**
Encode Transaction to json
 */
func (tx *Transaction) EncodeToJson() (string, error) {
	jsonTransaction, err := json.Marshal(tx)
	return string(jsonTransaction), err
}

/**
Encode TransactionData to json
 */
func (data *TransactionData) EncodeToJson() (string, error) {
	jsonTransactionData, err := json.Marshal(data)
	return string(jsonTransactionData), err
}
