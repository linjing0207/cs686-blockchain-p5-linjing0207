package p3

import (
	"../p1"
	"../p2"
	"../p5"
	"../p5/signature"
	"./data"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:6686"
var PREFIX = "000000"
var DEPOSIT = int32(100)

var SBC data.SyncBlockChain
var Peers data.PeerList
var Transactions p5.TransactionList
var ifStarted bool
var selfAddr string
var Port int32
var privateKey []byte
var publicKey []byte

/**
Pairs struct stores addr and id
 */
type Pairs struct {
	Addr string `json:"addr"`
	Id   int32  `json:"id"`
}

/**
Init(): Create SyncBlockChain and PeerList instances.
This function will be executed before everything else.
 */
func init() {
	// Do some initialization here.
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(Port, 32)
	//Transactions = data.NewTransactionList()
	ifStarted = false
}

/**
Start(): Get an ID from TA's server, download the BlockChain from your own first node,
use "go StartHeartBeat()" to start HeartBeat loop.
// Register ID, download BlockChain, start HeartBeat
Route{
		"Start",
		"GET",
		"/start",
		Start,
	}
 */
func Start(w http.ResponseWriter, r *http.Request) {

	if ifStarted == false {
		ifStarted = true
		//Get an ID from TA's server
		//Register()
		Peers.Register(Port)
		selfAddr = "http://localhost:" + strconv.Itoa(int(Port))
		var mpt p1.MerklePatriciaTrie
		if Port == 6688 { //node1
			//hard code original BlockChain for node1
			mpt = p1.MerklePatriciaTrie{}
			mpt.Initial()
		} else { //other nodes
			//download the BlockChain from your own first node
			Download()
			Peers.Add(TA_SERVER, 6688)
			mpt = SBC.GetLatestBlocks()[0].Value
		}
		//generate privateKey and publicKey
		signature.GenRsaKey(1024)
		privateKey = signature.GetPrivateKey()
		publicKey = signature.GetPublicKey()

		//create userInfo(balance + publicKey)
		userInfo := make(map[string]string)
		userInfo["balance"] = strconv.Itoa(int(DEPOSIT))
		userInfo["publicKey"] = string(publicKey)
		buffer, err := json.Marshal(userInfo)
		if err != nil {
			log.Println(err)
		}
		mpt.Insert(strconv.Itoa(int(Port)), string(buffer[:]))
		block := SBC.GenBlock(mpt, "", p5.TransactionData{})

		var peerMapJSON string
		peerMapJSON, err = Peers.PeerMapToJson()
		if err != nil {
			log.Println(err)
		}
		var blockJson string
		blockJson, err = block.EncodeToJson()
		heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJSON, selfAddr, blockJson)
		ForwardHeartBeat(heartBeatData)

		//start HeartBeat loop.
		go StartHeartBeat()
		//start StartTryingNonces loop
		go StartTryingNonces()
	}
	fmt.Fprintf(w, "start finish!\n")
}

/**
Show(): Shows the PeerMap and the BlockChain.
// Display peerList and sbc
Route{
		"Show",
		"GET",
		"/show",
		Show,
	}
 */
func Show(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

/**
Register(): Go to TA's server, get an ID.
After a new node is launched, it will go to "mc07.cs.usfca.edu:6688/peer" to register itself, and get an Id(nodeId).
*/
func Register() {
	//url := mc07.cs.usfca.edu:6688/peer
	url := REGISTER_SERVER

	resp, err := http.Get(url)
	if err != nil {
		log.Println("RegisterError: Get request", err)
		return
	}

	var id int
	if resp.StatusCode == http.StatusOK {
		var bodyBytes []byte
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("RegisterError: ReadAll", err)
			return
		}
		resp.Body.Close()
		bodyString := string(bodyBytes)
		id, err = strconv.Atoi(bodyString)
	}

	//setID to peerList
	Peers.Register(int32(id))
}

/**
Download(): Download the current BlockChain from your own first node(can be hardcoded).
Download blockchain from TA server
It's ok to use this function only after launching a new node.
You may not need it after node starts heartBeats.
 */
func Download() {

	//send request: upload
	//"http://localhost:6688/upload"
	url := BC_DOWNLOAD_SERVER

	//Every node(not node1) downloads BlockChain from node1(6688)
	pairs := Pairs{selfAddr, Port}
	pairsJson, err := json.Marshal(pairs)
	if err != nil {
		log.Println("DownloadMethodError:Marshal", err)
	}
	var resp *http.Response
	resp, err = http.Post(url, "application/json; charset=UTF-8", bytes.NewBuffer(pairsJson))
	if err != nil {
		log.Println("DownloadMethodError:Get request", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("DownloadMethodReadError:ReadAll", err)
	}
	resp.Body.Close()
	blockChainJson := string(body)
	//Equal to replace, because current node's BlockChain is empty
	SBC.UpdateEntireBlockChain(blockChainJson)
}

/**
Upload(): Return the BlockChain's JSON. And add the remote peer into the PeerMap.
Upload blockchain to whoever called this method, return jsonStr
Route{
		"Upload",
		"POST",
		"/upload",
		Upload,
	}
 */
func Upload(w http.ResponseWriter, r *http.Request) {
	//get bc from node1(TA server node)
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		log.Println("UploadMethodError: BlockChainToJson", err)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("HeartBeatReceive: ReadAll", err)
		httpResponseError(w, http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	pairs := Pairs{}
	err = json.Unmarshal([]byte(body), &pairs)
	if err != nil {
		log.Println("HeartBeatReceive: Umarshal failed", err)
		httpResponseError(w, http.StatusInternalServerError)
		return
	}

	//current node is node1: add peer
	//Every new node launched should let node1 know
	Peers.Add(pairs.Addr, pairs.Id)

	fmt.Fprint(w, blockChainJson+"\n")
}

/**
UploadBlock(): Return the Block's JSON.
Upload a block to whoever called this method, return jsonStr
If there's an error, return HTTP 500: InternalServerError.
Route{
		"UploadBlock",
		"GET",
		"/block/{height}/{hash}",
		UploadBlock,
	}
 */
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h, err := strconv.Atoi(vars["height"])
	if err != nil {
		log.Println("UploadBlockError: string to int", err)
		httpResponseError(w, http.StatusInternalServerError)
		return
	}
	height := int32(h)
	hash, found := vars["hash"]
	if !found {
		log.Println("UploadBlockError: cannot find hash", err)
		httpResponseError(w, http.StatusInternalServerError)
		return
	}
	var blockJson string
	block, blockExist := SBC.GetBlock(height, hash)
	if blockExist == true {
		// If you have the block, return the JSON string of the specific block;
		blockJson, err = block.EncodeToJson()
		if err != nil {
			log.Println("UploadBlockError: encodeToJson", err)
			//if there's an error, return HTTP 500: InternalServerError.
			httpResponseError(w, http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, blockJson+"\n")
	} else {
		//if you don't have the block, return HTTP 204: StatusNoContent;
		httpResponseError(w, http.StatusNoContent)
	}
}

/**
httpResponseError(): Write response header and content.
 */
func httpResponseError(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	responseContent := ""
	if statusCode == http.StatusInternalServerError {
		responseContent = "500 - InternalServerError!\n"
	} else if statusCode == http.StatusNoContent {
		responseContent = "204 - StatusNoContent!\n"
	}
	w.WriteHeader(statusCode)
	w.Write([]byte(responseContent))
}

/**
Details for HeartBeatReceive
1. When a node received a HeartBeat, the node will add the sender’s IP address,
along with sender’s PeerList into its own PeerList. At this time,the number of peers stored in PeerList might exceed 32 and it is ok.
As described in previously, you don’t have to rebalance every time you receive a HeartBeat.Rebalance happens only before you send HeartBeats.
2. If the HeartBeatData contains a new block, the node will first check if the previous block exists (the previous block is the block whose hash is the parentHash of the next block).
3. If the previous block doesn't exist, the node will ask every peer at "/block/{height}/{hash}" to download that block.
4. After making sure previous block exists, insert the block from HeartBeatData to the current BlockChain.
5. Since every node only has 32 peers, every peer will forward the new block to all peers according to its PeerList.
That is to make sure every user in the network would receive the new block.
For this project. Every HeartBeatData takes 2 hops, which means after a node received a HeartBeatData from the original block maker, the remaining hop times is 1.
*/
/**
HeartBeatReceive():
Steps:
Add the remote address, and the PeerMapJSON into local PeerMap.
Then check if the HeartBeatData contains a new block.
If so, do these: (1) check if the parent block exists.
If not, call AskForBlock() to download the parent block.
(2) insert the new block from HeartBeatData.
(3) HeartBeatData.hops minus one, and if it's still bigger than 0, call ForwardHeartBeat() to forward this heartBeat to all peers.
 Route{
		"HeartBeatReceive",
		"POST",
		"/heartbeat/receive",
		HeartBeatReceive,
	}
*/
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	//get hearBeatData
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("HeartBeatReceive: ReadAll", err)
		return
	}
	r.Body.Close()

	//heartBeatJson := string(body)
	heartBeatData := data.HeartBeatData{}
	err = json.Unmarshal([]byte(body), &heartBeatData)
	if err != nil {
		log.Println("HeartBeatReceive: Umarshal failed", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//peerList:
	//the node will add the sender’s IP address,along with sender’s PeerList into its own PeerList.
	if heartBeatData.Addr != selfAddr {
		Peers.Add(heartBeatData.Addr, heartBeatData.Id)
		Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, selfAddr)

		//Then check if the HeartBeatData contains a new block.
		if heartBeatData.IfNewBlock == true {
			block, _ := p2.DecodeFromJson(heartBeatData.BlockJson)
			//verify nonce
			nonceHash := GetNonceHash(block.Header.ParentHash, block.Header.Nonce, block.Value.GetRoot())
			if block.Header.Nonce == "" || strings.HasPrefix(nonceHash, PREFIX) {
				//check parentBlock
				//If parentBlock not exist in my bc, you should ask others to download that parent block of height 6 before inserting the block of height 7.
				if !SBC.CheckParentHash(block) {
					success := AskForBlock(block.Header.Height-1, block.Header.ParentHash)
					if success {
						parentBlock, _ := SBC.GetParentBlock(block)
						//check valid tx
						if block.Header.Transaction == (p5.TransactionData{}) ||
							validateTX(block.Header.Transaction, parentBlock) {
							SBC.Insert(block)
							Transactions.Delete(block.Header.Transaction)
						}
					}
					//else: failed -> forgive to insert whole chain
				} else { //parentBlock exists in my bc
					parentBlock, _ := SBC.GetParentBlock(block)
					//check valid tx
					if block.Header.Transaction == (p5.TransactionData{}) || validateTX(block.Header.Transaction, parentBlock) {
						SBC.Insert(block)
						Transactions.Delete(block.Header.Transaction)
					}
				}
			}
		}
		//hops-1
		heartBeatData.Hops -= 1 //initial hops = 3
		if heartBeatData.Hops > 0 {
			ForwardHeartBeat(heartBeatData)
		}
	}
	fmt.Fprintf(w, "HeartBeat recived.\n")
}

/**
AskForBlock(): Loop through all peers in local PeerMap to download a block.
AskForBlock(): Update this function to recursively ask for all the missing predesessor blocks instead of only the parent block.
As soon as one peer returns the block, stop the loop.
What to do:
Ask another server to return a block of certain height and hash
in AskForBlock you will call http get to /localhost:port/block/{height}/{hash} (UploadBlock) to get the Block
 */
//verify nonce
//recursive stop base case height = 1
func AskForBlock(height int32, hash string) bool {
	var block p2.Block
	if height >= 1 {
		//go to all peers, ask for block
		for addr := range Peers.Copy() {
			resp, err := http.Get(addr + "/block/" + string(height) + "/" + hash)
			if err != nil {
				log.Println("AskForBlockError: get request", err)
			}
			if resp.StatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println("AskForBlockError: ReadAll", err)
				}
				resp.Body.Close()

				blockJson := string(body)
				block, err = p2.DecodeFromJson(blockJson)
				if err != nil {
					log.Println("AskForBlockError: DecodeFromJson", err)
				}

				//verify nonce
				nonceHash := GetNonceHash(block.Header.ParentHash, block.Header.Nonce, block.Value.GetRoot())
				if block.Header.Nonce == "" || strings.HasPrefix(nonceHash, PREFIX) {
					//already get the block, now you have to check whether or not you have its' parent
					//parentBlock not exist in my bc, go to find parent's parent
					if !SBC.CheckParentHash(block) {
						success := AskForBlock(block.Header.Height-1, block.Header.ParentHash)
						if success {
							parentBlock, _ := SBC.GetParentBlock(block)
							//check valid tx
							if block.Header.Transaction == (p5.TransactionData{}) || validateTX(block.Header.Transaction, parentBlock) {
								SBC.Insert(block)
								Transactions.Delete(block.Header.Transaction)
								return true
							}
						}
					} else { // parentBlock exist in my bc, insert curBlock
						parentBlock, _ := SBC.GetParentBlock(block)
						//check valid tx
						if block.Header.Transaction == (p5.TransactionData{}) || validateTX(block.Header.Transaction, parentBlock) {
							SBC.Insert(block)
							Transactions.Delete(block.Header.Transaction)
							return true
						}
					}
				}
			}
		}
	}
	return false
}

/**
ForwardHeartBeat(): Send the HeartBeatData to all peers in local PeerMap.
After registration, the node will start to send HeartBeat for every 5~10 seconds.
Since every node only has 32 peers, every peer will forward the new block to all peers according to its PeerList.
That is to make sure every user in the network would receive the new block.
For this project. Every HeartBeatData takes 2 hops,
which means after a node received a HeartBeatData from the original block maker, the remaining hop times is 1.
*/
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	// The PeerList can temporarily hold more than 32 nodes,
	// Before sending HeartBeats, a node will first re-balance the PeerList by choosing the 32 closest peers.
	Peers.Rebalance()

	//send heartBeat to peers
	peerMap := Peers.Copy()
	var resp *http.Response
	for addr := range peerMap {
		url := addr + "/heartbeat/receive"
		heartBeatDataJson, err := json.Marshal(heartBeatData)
		if err != nil {
			log.Println("ForwardHeartBeatError: json.Marshal", err)
		}
		resp, err = http.Post(url, "application/json; charset=UTF-8", bytes.NewBuffer(heartBeatDataJson))
		if err != nil {
			log.Println("ForwardHeartBeatError: Post request", err)
		}
		if resp == nil {
			return
		}
		resp.Body.Close()
	}
}

/**
StartHeartBeat(): Start a while loop. Inside the loop, sleep for randomly 5~10 seconds, then use PrepareHeartBeatData() to create a HeartBeatData, and send it to all peers in the local PeerMap.
You can start with "Start", then "Send HeartBeat", then "Receive HeartBeat".
*/
func StartHeartBeat() {

	rand.Seed(time.Now().Unix())
	var myRand int
	//start the heartBeat loop.
	for {
		fmt.Println("time:", time.Now())

		peerMapJSON, err := Peers.PeerMapToJson()
		if err != nil {
			log.Println("StartHeartBeatError: PeerMapToJson", err)
		}
		heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJSON, selfAddr, "")

		//5~10s forward heartbeat
		//myRand = rand.Intn(11-5) + 5

		//5~10s forward heartbeat
		myRand = rand.Intn(15-5) + 5

		ForwardHeartBeat(heartBeatData)

		//sleep
		time.Sleep(time.Duration(myRand) * time.Second)
	}
}

/**
StartTryingNonces(): This function starts a new thread that tries different nonces to generate new blocks. Nonce is a string of 16 hexes such as "1f7b169c846f218a". Initialize the rand when you start a new node with something unique about each node, such as the current time or the port number. Here's the workflow of generating blocks:
    (1) Start a while loop.
    (2) Get the latest block or one of the latest blocks to use as a parent block.
    (3) Create an MPT.
    (4) Randomly generate the first nonce, verify it with simple PoW algorithm to see if SHA3(parentHash + nonce + mptRootHash) starts with 10 0's (or the number you modified into). Since we use one laptop to try different nonces, six to seven 0's could be enough. If the nonce failed the verification, increment it by 1 and try the next nonce.
    (6) If a nonce is found and the next block is generated, forward that block to all peers with an HeartBeatData;
    (7) If someone else found a nonce first, and you received the new block through your function ReceiveHeartBeat(), stop trying nonce on the current block, continue to the while loop by jumping to the step(2).
 */
//solve puzzle
func StartTryingNonces() {
	rand.Seed(time.Now().Unix())
	var nonce string
	//(1) Start a while loop.
	for {
		//fmt.Println("hello???:", Transactions.TxList)
		if len(Transactions.TxList) >= 1 {

			//(2) Get the latest block or one of the latest blocks to use as a parent block.
			parentBlock := SBC.GetLatestBlocks()[0]
			mpt := p1.MerklePatriciaTrie{}
			mpt = parentBlock.Value
			//get TX(highest TX fee) from pool:
			Transactions.SortByTxFee()
			//get last one = with highest TX fee
			tx := Transactions.TxList[len(Transactions.TxList)-1]

			//this tx is data without hops
			txData := tx.TransactionData

			curPayer := strconv.Itoa(int(txData.Msg.PayerId))
			curPayerBalance, _ := getBalanceAndPk(mpt, curPayer)

			curPayee := strconv.Itoa(int(txData.Msg.PayeeId))
			curPayeeBalance, _ := getBalanceAndPk(mpt, curPayee)
			//verify!!!
			//check payee exist && payer has enough balance
			//curPayer != curPayee &&
			if curPayer != curPayee && curPayeeBalance >= 0 && curPayerBalance >= txData.Msg.Amount {

				//(3) Create an MPT.
				//Deduct balance from current TX payer
				mpt = updateBalance(txData.Msg.PayerId, -(txData.Msg.Amount + txData.Msg.TxFee), mpt)

				curHeight := SBC.GetLength()              //10, next 11
				blockList, bool := SBC.Get(curHeight - 5) //5
				if bool == true {
					var prevBlock p2.Block
					//go back to the sixth block in front of current block
					for i := 0; i < 5; i++ {
						prevBlock, _ = SBC.GetParentBlock(parentBlock)
					}
					prevTX := prevBlock.Header.Transaction
					//check fork
					//has fork
					if len(blockList) > 1 {
						for j := range blockList {
							if prevBlock.Header.Hash != blockList[j].Header.Hash {
								forkTX := blockList[j].Header.Transaction
								//Add: back
								mpt = updateBalance(forkTX.Msg.PayerId, forkTX.Msg.Total, mpt)
								//Transactions.Add(tx)
							}
						}
					}
					//Add: miner get tx fee and payee get money
					mpt = updateBalance(prevTX.Msg.PayeeId, prevTX.Msg.Amount, mpt)
					mpt = updateBalance(prevTX.MinerId, prevTX.Msg.TxFee, mpt)
				}

				nonce = ""
				//(4) Randomly generate the first nonce
				for i := 0; i < 16; i++ {
					rand := strconv.FormatInt(int64(rand.Intn(16)), 16)
					nonce += rand
				}
				parentHash := parentBlock.Header.Hash
				mptRoot := mpt.GetRoot()

				for {
					//(7) If someone else found a nonce first,
					// and you received the new block through your function ReceiveHeartBeat(),
					// stop trying nonce on the current block,
					// continue to the while loop by jumping to the step(2).
					if parentBlock.Header.Height < SBC.GetLatestBlocks()[0].Header.Height {
						break
					} else {
						nonceHash := GetNonceHash(parentHash, nonce, mptRoot)
						if strings.HasPrefix(nonceHash, PREFIX) {
							//(6) If a nonce is found and the next block is generated,
							// forward that block to all peers with an HeartBeatData;
							txData.MinerId = Port
							block := SBC.GenBlock(mpt, nonce, txData)
							blockJson, err := block.EncodeToJson()
							if err != nil {
								log.Println("StartTryingNonces: EncodeToJson", err)
							}
							peerMapJSON, err := Peers.PeerMapToJson()
							if err != nil {
								log.Println("StartTryingNonces: PeerMapToJson", err)
							}
							heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJSON, selfAddr, blockJson)
							fmt.Println("Generate block:")

							//forward heartbeat
							ForwardHeartBeat(heartBeatData)
							//remove tx from pool
							Transactions.Delete(tx.TransactionData)
							break
						} else {
							data, err := strconv.ParseUint(nonce, 16, 64)
							if err != nil {
								fmt.Println(err)
							}
							nonce = strconv.FormatInt(int64(data+1), 16)
							if len(nonce) > 16 {
								nonce = "0000000000000000"
							}
						}
					}
				}
			} else {
				Transactions.Delete(tx.TransactionData)
			}
		}
	}
}

/**
update balance in mpt, return back mpt
 */
func updateBalance(id int32, amount int32, mpt p1.MerklePatriciaTrie) p1.MerklePatriciaTrie {
	user := strconv.Itoa(int(id))
	oldBalance, pk := getBalanceAndPk(mpt, user)
	newBalance := strconv.Itoa(int(oldBalance + amount))

	infoMap := make(map[string]string)
	infoMap["balance"] = newBalance
	infoMap["publicKey"] = pk
	buffer, err := json.Marshal(infoMap)
	if err != nil {
		fmt.Println(err)
	}
	mpt.Insert(user, string(buffer[:]))
	return mpt
}


func stringToInt32(str string) int32 {
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Println("stringToInt32: convertStringToInt32", err)
	}
	return int32(v)
}

/**
By using key to find the balance and public key
 */
func getBalanceAndPk(mpt p1.MerklePatriciaTrie, key string) (int32, string) {
	value, err := mpt.Get(key)
	if err != nil {
		log.Println("getBalance: Get", err)
		return -1, ""
	}
	infoMap := make(map[string]string)
	err = json.Unmarshal([]byte(value), &infoMap)
	if err != nil {
		fmt.Println(err)
	}
	var b int
	b, err = strconv.Atoi(infoMap["balance"])
	p := infoMap["publicKey"]
	return int32(b), p
}

/**
validate tx
 */
func validateTX(tx p5.TransactionData, parentBlock p2.Block) bool {
	payer := strconv.Itoa(int(tx.Msg.PayerId))
	payee := strconv.Itoa(int(tx.Msg.PayeeId))
	payerBalance, _ := getBalanceAndPk(parentBlock.Value, payer)
	payeeBalance, _ := getBalanceAndPk(parentBlock.Value, payee)
	//valid
	if payer != payee && payeeBalance >= 0 && payerBalance >= tx.Msg.Amount {
		return true
	} else {
		return false
	}
}

/**
GetNonceHash: use SHA3(parentHash + nonce + mptRootHash) to get nonce hash
 */
func GetNonceHash(parentHash string, nonce string, mptRoot string) string {
	hashStr := parentHash + nonce + mptRoot
	sum := sha3.Sum256([]byte(hashStr))
	nonceHash := hex.EncodeToString(sum[:])
	return nonceHash
}

/**
//send TX to node1, node1 will send TX to other peers
*/
/**
Route{
		"Transfer",
		"POST",
		"/transfer",
		Transfer,
	},
json:
{
	"PayeeId": 6688,
	"Amount": 20,
	"TxFee": 5
}
 */
func Transfer(w http.ResponseWriter, r *http.Request) {
	//get transaction
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Transfer: ReadAll", err)
		return
	}
	//body -> tx
	//double check
	txPost := p5.TransactionPost{}
	err = json.Unmarshal([]byte(body), &txPost)
	r.Body.Close()
	if err != nil {
		log.Println("Transfer: Umarshal failed", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	msg := p5.NewMsg(Port, txPost.PayeeId, txPost.Amount, txPost.TxFee, txPost.Amount+txPost.TxFee)

	tx := addSignature(msg)
	Transactions.Add(tx)
	ForwardTransaction(tx)

	fmt.Fprintf(w, "Create TX success.\n")

}

/**
add signature for msg and return back tx
 */
func addSignature(msg p5.Msg) p5.Transaction {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("original jsonMsg:", string(jsonMsg))
	//use private key to sign
	sig, _ := signature.RsaSign([]byte(jsonMsg))
	//fmt.Println("test:", hex.EncodeToString(sig))
	//fmt.Println("original signature:    ", hex.EncodeToString(sig))
	//verify signature
	//fmt.Println(signature.RsaSignVer([]byte(jsonMsg), sig))

	txData := p5.TransactionData{msg, -1, hex.EncodeToString(sig)}
	tx := p5.NewTransaction(Port, txData, 3)
	//var jsonTX []byte
	//jsonTX, err = json.Marshal(tx)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("jsonText:", hex.EncodeToString(jsonTX))
	return tx

}



/**
Route{
		"TransactionReceive",
		"POST",
		"transaction/receive",
		TransactionReceive,
	},
 */
func TransactionReceive(w http.ResponseWriter, r *http.Request) {
	//get transaction
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("TransactionReceive: ReadAll", err)
		return
	}
	r.Body.Close()

	//heartBeatJson := string(body)
	tx := p5.Transaction{}
	err = json.Unmarshal([]byte(body), &tx)
	if err != nil {
		log.Println("TransactionReceive: Umarshal failed", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if tx.Id != Port {
		//add TX to pool
		Transactions.Add(tx)
		//hops-1
		tx.Hops -= 1 //initial hops = 3
		//Hops check
		if tx.Hops > 0 {
			ForwardTransaction(tx)
		}
	}
}

/**
verify signature of tx
 */
func verifySignature(jsonTX string) bool {
	//send json to others
	t := p5.Transaction{}
	err := json.Unmarshal([]byte(jsonTX), &t)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("t:", t)
	//fmt.Println(hex.EncodeToString(sig))
	//get signature
	sign, _ := hex.DecodeString(t.TransactionData.Signature)
	msg1, _ := t.TransactionData.Msg.EncodeToJson()
	//fmt.Println("after    signature:    ", sign)
	//fmt.Println("after    jsonMsg:", msg1)
	//verify signature
	//fmt.Println(signature.RsaSignVer([]byte(msg1), sign))
	if signature.RsaSignVer([]byte(msg1), sign) == nil {
		return true
	} else {
		return false
	}
}

/**
forward transaction to other miner
 */
func ForwardTransaction(tx p5.Transaction) {
	Peers.Rebalance()
	//send tx to peers
	peerMap := Peers.Copy()
	var resp *http.Response
	for addr := range peerMap {
		url := addr + "/transaction/receive"
		txJson, err := json.Marshal(tx)
		if err != nil {
			log.Println("ForwardTransaction: json.Marshal", err)
		}
		resp, err = http.Post(url, "application/json; charset=UTF-8", bytes.NewBuffer(txJson))
		if err != nil {
			log.Println("ForwardTransaction: Post request", err)
		}
		if resp == nil {
			return
		}
		resp.Body.Close()
	}
}

/**
Canonical(): This function prints the current canonical chain, and chains of all forks if there are forks.
Note that all forks should end at the same height (otherwise there wouldn't be a fork).
Route{
		"Canonical",
		"GET",
		"/canonical",
		Canonical,
	},
 */
func Canonical(w http.ResponseWriter, r *http.Request) {
	str := ""
	blockList := SBC.GetLatestBlocks()
	//fmt.Println(blockList)
	for i, block := range blockList {
		str += "\n" + "CHAIN #" + strconv.Itoa(i+1) + "\n"
		str += getBlockFormat(block) + "\n"
		parentBlock, parentBlockExist := SBC.GetParentBlock(block)
		for parentBlockExist { //find next parentBlock
			str += getBlockFormat(parentBlock) + "\n"
			parentBlock, parentBlockExist = SBC.GetParentBlock(parentBlock)
		}
	}
	fmt.Fprintf(w, str)
}
/**
get balance on current user account
Route{
		"MyBalance",
		"GET",
		"/mybalance",
		MyBalance,
	},
 */
func MyBalance(w http.ResponseWriter, r *http.Request) {
	//Assume 0th is the canonical
	block := SBC.GetLatestBlocks()[0]
	//blockJson, _ := block.EncodeToJson()
	//fmt.Println("block:", blockJson)
	var mpt p1.MerklePatriciaTrie
	mpt = block.Value
	fmt.Println("mpt:", mpt)
	balance, pk := getBalanceAndPk(mpt, strconv.Itoa(int(Port)))
	str := "SelfId:" + strconv.Itoa(int(Port)) + "\n"
	str += "Balance:" + strconv.Itoa(int(balance)) + "\n"
	str += "PublicKey:\n" + pk + "\n"
	fmt.Fprintf(w, str)
}

/**
get all my txs
Route{
		"MyTXs",
		"GET",
		"/mytxs",
		MyTXs,
	},
 */
func MyTXs(w http.ResponseWriter, r *http.Request) {
	str := ""
	blockList := SBC.GetLatestBlocks()
	for i, block := range blockList {
		str += "\n" + "CHAIN #" + strconv.Itoa(i+1) + "\n"
		str += getTX(block)
		parentBlock, parentBlockExist := SBC.GetParentBlock(block)
		for parentBlockExist { //find next parentBlock
			str += getTX(parentBlock)
			parentBlock, parentBlockExist = SBC.GetParentBlock(parentBlock)
		}
	}
	fmt.Fprintf(w, str)
}

/**
get tx from a block
 */
func getTX(block p2.Block) string {

	if block.Header.Transaction.Msg.PayerId == Port || block.Header.Transaction.Msg.PayeeId == Port {
		txDetials, _ := block.Header.Transaction.EncodeToJson()
		return txDetials + "\n"
	} else {
		return ""
	}
}

/**
This function prints all TXs
Route{
		"AllTXs",
		"GET",
		"/txs",
		AllTXs,
	},
 */
func AllTXs(w http.ResponseWriter, r *http.Request) {
	str := "txs:\n"
	for i := range Transactions.TxList {
		tx, err := Transactions.TxList[i].EncodeToJson()
		if err != nil {
			fmt.Println(err)
		}
		str += tx + "\n"
	}
	fmt.Fprintf(w, str)
}

/**
getBlockFormat: create block format
 */
func getBlockFormat(block p2.Block) string {
	return "height=" + strconv.Itoa(int(block.Header.Height)) + ", timestamp=" + strconv.Itoa(int(block.Header.Timestamp)) + ", hash=" + block.Header.Hash + ", parentHash=" + block.Header.ParentHash + ", size=" + strconv.Itoa(int(block.Header.Size))
}
