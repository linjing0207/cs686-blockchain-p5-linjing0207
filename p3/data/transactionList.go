package data

import (
	"fmt"
	"sort"
	"sync"
)

type TransactionList struct {
	//selfId int32
	TxList []TransactionData
	mux    sync.Mutex
}

type TransactionData struct {
	PayerId int32
	PayeeId int32
	Amount  int32
	TxFee   int32
	Total   int32
	//Hops    int32 `json:"hops"`
	//PublicKey string
}

func NewTransactionList() TransactionList {
	//??
	txList := TransactionList{TxList: []TransactionData{}}
	return txList
}

func (txs *TransactionList) Add(tx TransactionData) {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	//same transaction??
	for i := range txs.TxList {
		if txs.TxList[i] == tx {
			return
		}
	}
	txs.TxList = append(txs.TxList, tx)
}

//func (txs *TransactionList) TXEncodeToJson (string, error){
//
//	var json string
//	buffer, err := json.Marshal(txs.TxList)
//	if err != nil {
//		log.Println(err)
//	}
//
//	//fmt.Println(string(buffer[:]))
//	json = string(buffer[:])
//
//	return json, err
//}

func (txs *TransactionList) Delete(tx TransactionData) {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	s := txs.TxList
	for i := range s {
		if s[i] == tx {
			//exchange the last one
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
		}
	}
}

func (txs *TransactionList) SortByTxFee() {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	sort.Slice(txs.TxList, func(i, j int) bool { return txs.TxList[i].TxFee < txs.TxList[j].TxFee })
	fmt.Println("By TX fee:", txs.TxList)
	//sort function
}
