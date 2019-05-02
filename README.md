# Progress report for Project5- Electronic Funds Transfer System

## Introduction:
> [A cryptocurrency (or crypto currency) is a digital asset designed to work as a medium of exchange that uses strong cryptography to secure financial transactions, control the creation of additional units, and verify the transfer of assets.](https://en.wikipedia.org/wiki/Cryptocurrency)
This project implements a cryptocurrency system based on the project4 structure. In this project design, we assume that all the users could be miners. That means all the nodes could transfer money to others and also could generate blocks.

## Background:

> Blockchain technology could avoid building complex systems, and create a more direct payment process between the payer and the payee. Whether it is domestic transfer or cross-border transfer, this payment has the advantages of low price, fast, and no intermediate fee.

## Functionalities:

> 1.	Each block stores all user’s newest balance and includes each TX happened in that time. [4/18]
> 2.	API: show balance and transactions. [4/20]
> 3.	User could transfer ether coins to other users. [4/22]
> 4.	Miners add TX to mempool and forward TX to peers. [4/22]
> 5.	Miner will priority serves the transactions with high TX fee, validate TX. [4/25]
> 6.	When miner generate a new block, miner will process transaction and get TX fee, then forward heartbeat to peers. [4/28]
> 7.	When receiving a new block, verify nonce and validate the TX. [5/1]
> 8.	Money will be refunded when transaction fails (when block becomes a fork). [5/4]
> 9.	Miners can earn TX fee, payee could get money when transaction be confirmed. (after 6 blocks) [5/8]
> 10.	Each transaction will include payer’s signature. [5/12]
> 11.	Final Testing [5/14]

## What you accomplished now and how:

> Note: I have already accomplished functionalities from 1 to 7.
> ### 1.	Data structure modification:
> #### In Block.go:
>> **Block**:	Block{Header{Height, Timestamp, Hash, ParentHash, Size, Transaction}, Value}<br>
Each block must contain a header, and in the header, add a transaction field based on previous structure. <br>
Each block must have a value, which is a Merkle Patricia Trie. All the data are inserted in the MPT and then a block contains that MPT as the value. MPT stores user’s ID and their account balance.<br>
Value: mpt MerklePatriciaTrie<br>
> #### In transactionList.go:
>> **TransactionList**:	TransctionList {TxList, mux}<br>
(1)	TxList: [] TransctionData<br>
(2)	mux(lock)<br>
>> **TransctionData**:  TransctionData {PayId, PayeeId, Amount, TxFee, Total}<br>
Transaction stores TX details.<br>
(1)	PayId: int32<br>
(2)	PayeeId: int32<br>
(3)	Amount: int32<br>
(4)	TxFee: int32<br>
(5)	Total: int32<br>

### 2.	API

> /transaction/receive<br>
Method: POST<br>
Request: TransactionData<br>
Description: Receive a transaction<br>
Logic: Miner receive TX, add to his mempool, then forward to other miners.<br>

> /mybalance<br>
Method: GET<br>
Response: Account ID and balance.<br>
Description: Show the balance of account.<br>
Logic: Get the value of latest block, search the balance by ID in MPT, return Id and balance.<br>

> /mytxs<br>
Method: GET<br>
Response: The JSON string of all the transactions.<br>
Description: Show all the transactions of current user.<br>
Logic: Go through the canonical chain, if payer or payee is current user, return back the list of TXs.<br>

### 3.	Functionalities implementation

> (1)	User could transfer ether coins to other users.<br>
By using API “/transaction/receive”, user could send a transaction with TX fee to miners.

> (2)	Miners add TX to mempool and forward TX to peers.<br>
When Miner received TX, he will add this transaction to his own unconfirmed transaction list and forward TX to peers. Forward TX also using API “/transaction/receive” to send TX to other miners.

> (3)	Miner will priority serves the transactions with high TX fee, validate TX.<br>
Miner will try nonce constantly. Once he finds the nonce, he will sort all TX by TX fee in the mempool and get the TX with highest TX fee due to miners want to maximize their income. After that, miner have to validate the TX details. Check if there is enough balance for payer and check if the payee exists based on the latest block. 

> (4)	When miner generate a new block, miner will process transaction and get TX fee, then forward heartbeat to peers.<br>
Miner will generate the block for the valid transaction, and transfer money from payer to payee. At the same time, miner get the transaction fee. Finally, miner send heartbeat with the new block to his peers.

> (5)	When receiving a new block, verify nonce and validate the TX.<br>
When receiving a heartbeat with new block, user have to find all missing predecessor block from peers. Verify the nonce and validate the TX for each block before insert.

## Reference:

> 1.https://en.wikipedia.org/wiki/Cryptocurrency<br>
> 2.https://www.bitcoinmining.com/bitcoin-mining-fees/<br>
> 3.https://hbr.org/2017/01/the-truth-about-blockchain<br>

