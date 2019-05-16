# Final Report - Electronic Funds Transfer System

## Introduction:
> [A cryptocurrency (or crypto currency) is a digital asset designed to work as a medium of exchange that uses strong cryptography to secure financial transactions, control the creation of additional units, and verify the transfer of assets.](https://en.wikipedia.org/wiki/Cryptocurrency)
This project implements a cryptocurrency system based on the project4 structure. In this project design, we assume that all the users could be miners. That means all the nodes could transfer money to others and also could generate blocks.

## Background:

> Blockchain technology could avoid building complex systems, and create a more direct payment process between the payer and the payee. Whether it is domestic transfer or cross-border transfer, this payment has the advantages of low price, fast, and no intermediate fee.

## Functionalities(Acturally Accomplished Date):

>1.	Each block stores all user’s latest balance, public key, and includes each TX happened in that time. [4/18]
>2.	Each user will have 100 ETH deposits when he registers. [4/18]
>3.	API: show balance and transactions. [4/20]
>4.	User could transfer ether coins to other users, send TX to miners. [4/22]
>5.	Miners add TX to mempool and forward TX to peers. [4/22]
>6.	Miner will priority serves the transactions with high TX fee, validate TX. [4/25]
>7.	When miner generate a new block, miner will process a part of transaction (deduct money from payer), then forward heartbeat to peers. [4/28]
>8.	When receiving a new block, verify nonce and validate the TX. [5/1]
>9.	Miners can earn TX fee, payee could get money when transaction be confirmed. (after 6 blocks) [5/4]
>10.	Money will be refunded when transaction fails (when block becomes a fork). [5/8]
>11.	Payer will sign transaction before sending to miner and use public key to verify. [5/12]
>12.	Final Testing [5/16]


## What you accomplished now and how:

### 1.	Data structure modification:
> In Block.go:
>> **Block**:	Block{Header{Height, Timestamp, Hash, ParentHash, Size, Transaction}, Value}<br>
Each block must contain a header, and in the header, add a transaction field based on previous structure. <br>
Each block must have a value, which is a Merkle Patricia Trie. All the data are inserted in the MPT and then a block contains that MPT as the value. MPT stores user’s ID and their account balance.<br>
Value: mpt MerklePatriciaTrie<br>

> In transactionList.go:
>> **TransactionList**:	TransctionList {TxList, mux}<br>
(1)	TxList: [] TransctionData<br>
(2)	mux(lock)<br>

> In transaction.go:
>> **Transaction**:  Transaction {Id, TransactionData, Hops}<br>
Transaction stores payerId, TX content and hops.<br>
(1)	Id: int32<br>
(2)	TransactionData: TransactionData<br>
(3)	Hops: int32<br>

>> **TransactionData**:  TransactionData {Msg, MinerId, Signature}<br>
TransactionData data stores message and signatuer.<br>
(1)	Msg: Msg<br>
(2)	MinerId: int32<br>
(3)	Signature: string<br>

>> **Msg**:  Msg {PayerId, PayeeId, Amount, TxFee, Total}<br>
Msg stores TX details.<br>
(1)	PayerId: int32<br>
(2)	PayeeId: int32<br>
(3)	Amount: int32<br>
(4)	TxFee: int32<br>
(5)	Total: int32<br>

>> **TransactionPost**:  TransactionPost {PayeeId, Amount, TxFee}<br>
TransactionPost is used to convert post request body to struct.<br>
(1)	PayeeId: int32<br>
(2)	Amount: int32<br>
(3)	TxFee: int32<br>


### 2.	API
Note: Based on project4, the following are new API for project5.

> /transfer<br>
Method: POST<br>
Request: TransactionPost<br>
Description: A transfer request.<br>
Example request body:
{"PayeeId": 6688,"Amount": 20,"TxFee": 5}<br>
Logic: User send tx to miners.<br>

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
Logic: Go through the canonical chain, if payer or payee is current user, return back TXs.<br>

> /txs<br>
Method: GET<br>
Response: The list of all uncomfirmed transactions.<br>
Description: Show all the transactions in current user's mempool.<br>
Logic: Traverse transactionList, print out all uncomfirmed transactions.<br>

### 3.	Functionalities implementation
>(1) Data storage and initial deposit.
See data structure details. We assume that each user will have 100 ETH deposits when registering. What is more, user will get a pair of private key and public key. The first node will create a mpt to store its’ information (key is id, value is a map including balance and public key). Other node will update the previous mpt (insert its’ information). Then generating a new block without TX, and forwarding heartbeat to peers.

>(2) User could transfer ether coins to other users, send TX to miners.
This function implemented by API “/transfer”, with request body {"PayeeId": 6688,"Amount": 20,"TxFee": 5}. Add field payerId, total to the existing information, it could compose a message.

>(3) Each transaction will include payer’s signature.
User will get a private key after register. When user trying to post a transaction, he have to sign this transaction before sending to miners. 

>(4) Miners add TX to mempool and forward TX to peers.
API “/transaction/receive”: when miner received a TX, he will add this transaction to his own menpool (unconfirmed transaction list) and forward TX to peers. 

>(5) Miner will priority serves the transactions with high TX fee, validate TX.
Miner will try nonce constantly. They will sort all TXs by TX fee from the mempool and serve the TX with highest TX fee due to miners want to maximize their income. After that, miner have to validate the TX details. Check validation. 1.User could not transfer from to himself. 2.If there is enough balance for payer to transfer. 3.Check if the payee exists based on the latest block. What is more, miner are supposed to verify the signature of TX. Equally important that miner should verify the signature before generating a new block.

>(6) When miner generate a new block, miner will process transaction and forward heartbeat to peers.
Once miner finds the nonce, miner will generate the block for the valid transaction, and deduct money from payer. Then, miner send heartbeat with the new block to his peers.

>(7) When receiving a new block, verify nonce and validate the TX.
When receiving a heartbeat with new block, user have to find all missing predecessor block from peers. Verify the nonce and validate the TX for each block before insertion. After insertion, they must remove the corresponding TX from their mempool.

>(8) Miners can earn TX fee, payee could get money when transaction be confirmed. 
When 5 valid blocks have been generated after one block, that means this block has been confirmed. At the same time, miner who are generating the new block (the sixth block after this one) will process the previous transaction. He will put the transaction amount into payee’s account and put the transaction fee into previous miner’s account.

>(9) Money will be refunded when transaction fails (when block becomes a fork).
When a block has been confirmed, remaining blocks with same height became forks. At the same time, miner who are generating the new block (the sixth block after this one) will refund the money to previous payer, and put this transaction back to mempool again.


## Reference:

> 1.https://en.wikipedia.org/wiki/Cryptocurrency<br>
> 2.https://www.bitcoinmining.com/bitcoin-mining-fees/<br>
> 3.https://hbr.org/2017/01/the-truth-about-blockchain<br>

