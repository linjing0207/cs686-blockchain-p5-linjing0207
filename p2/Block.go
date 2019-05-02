package p2

import (
	"../p1"
	"../p3/data"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
)

/**
Each block must contain a header, and in the header there are the following fields:
(1) Height: int32
(2) Timestamp: int64
The value must be in the UNIX timestamp format such as 1550013938
(3) Hash: string.
(4) ParentHash: string
(5) Size: int32
Value: mpt MerklePatriciaTrie
Here's the summary of block structure:
The size is the length of the byte array of the block value
Block: Block{Header{Height, Timestamp, Hash, ParentHash, Size}, value}
 */
type Block struct {
	Header Header                `json:"header"`
	Value  p1.MerklePatriciaTrie `json:"mpt"`
}

/**
Header：
*/
type Header struct {
	Height      int32                `json:"height"`
	Timestamp   int64                `json:"timestamp"`
	Hash        string               `json:"hash"`
	ParentHash  string               `json:"parenthash"`
	Size        int32                `json:"size"`
	Nonce       string               `json:"nonce"`
	Transaction data.TransactionData `json:"transaction"`
}

/**
BlockJson is a struct for storing json format of the block
 */
type BlockJson struct {
	Height      int32                `json:"height"`
	Timestamp   int64                `json:"timeStamp"`
	Hash        string               `json:"hash"`
	ParentHash  string               `json:"parentHash"`
	Size        int32                `json:"size"`
	MPT         map[string]string    `json:"mpt"`
	Nonce       string               `json:"nonce"`
	Transaction data.TransactionData `json:"transaction"`
}

/**
Create a new block
Return type: Block
*/
func NewBlock(height int32, timeStamp int64, parentHash string, value p1.MerklePatriciaTrie, nonce string, tx data.TransactionData) Block {
	//create a block structure, value is null
	block := Block{}
	block.Initial(height, timeStamp, parentHash, value, nonce, tx)
	return block
}

/**
Description: This function takes arguments(such as height, parentHash, and value of MPT type) and forms a block.
This is a method of the block struct.
Block: Block{Header{Height, Timestamp, Hash, ParentHash, Size}, value}
Argument: height, timeStamp, hash, parentHash, value(mpt type)
 */
func (b *Block) Initial(height int32, timeStamp int64, parentHash string, value p1.MerklePatriciaTrie, nonce string, tx data.TransactionData) {
	//The size is the length of the byte array of the block value
	size := len(value.MptToByteArray())
	b.Header = Header{height, timeStamp, "", parentHash, int32(size), nonce, tx}
	//!!!create a header without hash first, then set hash(call hashBlock method)
	b.Header.Hash = b.hashBlock()
	b.Value = value
}

/**
Description: This function takes a string that represents the JSON value of a block as an input,
and decodes the input string back to a block instance.
Note that you have to reconstruct an MPT from the JSON string, and use that MPT as the block's value.
Argument: a string of JSON format
Return value: a block instance, error
 */
func DecodeFromJson(jsonString string) (Block, error) {

	blockJson := BlockJson{}
	err := json.Unmarshal([]byte(jsonString), &blockJson)
	if err != nil {
		fmt.Println("Umarshal failed:", err)
	}
	block := blockJsonToBlock(blockJson)

	return block, err
}

/**
Description: This function convert BlockJson struct to a block instance。
Use mpt paris to create mpt, then create block.
Argument: BlockJson
Return value: Block
 */
func blockJsonToBlock(blockJson BlockJson) Block {
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()

	mptMap := blockJson.MPT
	for k, v := range mptMap {
		mpt.Insert(k, v)
	}

	height := blockJson.Height
	timeStamp := blockJson.Timestamp
	hash := blockJson.Hash
	parentHash := blockJson.ParentHash
	size := blockJson.Size
	nonce := blockJson.Nonce
	tx := blockJson.Transaction

	header := Header{height, timeStamp, hash, parentHash, size, nonce, tx}
	block := Block{header, mpt}

	return block
}

/**
Description: This function encodes a block instance into a JSON format string.
Note that the block's value is an MPT, and you have to record all of the (key, value) pairs that have been inserted into the MPT in your JSON string.
Argument: a block or you may define this as a method of the block struct
Return value: a string of JSON format, error
Example：
{
    "hash":"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48",
    "timeStamp":1234567890,
    "height":1,
    "parentHash":"genesis",
    "size":1174,
    "mpt":{
        "charles":"ge",
        "hello":"world"
    }
}
*/
func (b *Block) EncodeToJson() (string, error) {
	var jsonBlock string
	blockJson := b.blockToBlockJson()

	buffer, err := json.Marshal(blockJson)
	if err != nil {
		log.Println(err)
	}

	//fmt.Println(string(buffer[:]))
	jsonBlock = string(buffer[:])

	return jsonBlock, err
}

/**
Description: This function convert block instance to a BlockJson struct。
Because the value of block is a mpt type, use root to traverse mpt to get all pairs.
Return value: BlockJson
 */
func (b *Block) blockToBlockJson() BlockJson {
	blockJson := BlockJson{}

	//traverse and find all the pairs in mpt
	//map{"charles":"ge","hello":"world"}
	mptMap := b.Value.GetMptMap(b.Value.GetRoot(), []uint8{})

	blockJson.Hash = b.Header.Hash
	blockJson.Timestamp = b.Header.Timestamp
	blockJson.Height = b.Header.Height
	blockJson.ParentHash = b.Header.ParentHash
	blockJson.Size = b.Header.Size
	blockJson.MPT = mptMap
	blockJson.Nonce = b.Header.Nonce
	blockJson.Transaction = b.Header.Transaction

	return blockJson
}

/**
Block’s hash is the SHA3-256 encoded value of this string(note that you have to follow this specific order):
hash_str := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + b.Value.Root + string(b.Header.Size)
Return: string
 */
func (b *Block) hashBlock() string {
	hashStr := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + b.Value.GetRoot() + string(b.Header.Size)
	sum := sha3.Sum256([]byte(hashStr))
	return hex.EncodeToString(sum[:])
}
