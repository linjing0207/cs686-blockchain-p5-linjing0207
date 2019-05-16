package p5

import (
	"sort"
	"sync"
)

type TransactionList struct {
	//selfId int32
	TxList []Transaction
	mux    sync.Mutex
}


/**
new transaction list
 */
func NewTransactionList() TransactionList {
	//??
	txList := TransactionList{TxList: []Transaction{}}
	return txList
}

/**
add tx into list
 */
func (txs *TransactionList) Add(tx Transaction) {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	//same transaction??
	for i := range txs.TxList {
		if txs.TxList[i].TransactionData.Msg == tx.TransactionData.Msg {
			//fmt.Println("没加")
			return
		}
	}
	txs.TxList = append(txs.TxList, tx)
}


/**
remove tx from list
 */
func (txs *TransactionList) Delete(txData TransactionData) {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	s := txs.TxList
	for i := range s {
		if s[i].TransactionData.Msg == txData.Msg {
			//exchange the last one
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
		}
	}
	txs.TxList = s
	//fmt.Println(txs.TxList)
}

/**
sort transactionList by tx fee
 */
func (txs *TransactionList) SortByTxFee() {
	txs.mux.Lock()
	defer txs.mux.Unlock()
	if len(txs.TxList) > 1 {

		sort.Slice(txs.TxList, func(i, j int) bool {
			return txs.TxList[i].TransactionData.Msg.TxFee < txs.TxList[j].TransactionData.Msg.TxFee
		})
	}
	//fmt.Println("By TX fee:", txs.TxList)
}
