package data

import (
	"../../p1"
	"../../p2"
	"sync"
	"time"
)

type SyncBlockChain struct {
	bc p2.BlockChain
	mux sync.Mutex
}

/**
Create a SyncBlockChain instance
 */
func NewBlockChain() SyncBlockChain {

	return SyncBlockChain{bc: p2.NewBlockChain()}
}

/**
Get blockList of a certain height from blockchain
 */
func(sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	var blockListExist = false
	blockList := sbc.bc.Get(height)
	if blockList != nil {
		blockListExist = true
	}
	return blockList, blockListExist
}

/**
Get a certain block
 */
func(sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	var block p2.Block
	var blockExist = false
	blockList := sbc.bc.Get(height)
	if blockList != nil {
		for _, v := range blockList {
			if v.Header.Hash == hash {
				block = v
				blockExist = true
				break
			}
		}
	}
	return block, blockExist
}

/**
Insert block to SyncBlockChain
 */
func(sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

/**
CheckParentHash(): This function would check if the block with the given "parentHash" exists in the blockChain.
If we have the parent block, we can insert the next block; if we don't have the parent block, we have to download the parent block before inserting the next block.
 */
func(sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	var parentHashExist = false
	parentHash := insertBlock.Header.ParentHash
	//previous height
	height := insertBlock.Header.Height - 1
	blockList := sbc.bc.Get(height)
	if blockList != nil {
		for _, v := range blockList {
			if v.Header.Hash == parentHash {
				parentHashExist = true
				break
			}
		}
	}
	return parentHashExist
}


/**
Replace my blockChain to peer's blockChain
 */
func(sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	peerBlockChain, _ := p2.DecodeJsonToBlockChain(blockChainJson)
	sbc.bc = peerBlockChain

	//merge my bc with peer's bc
	//for height, peerBlockList := range peerBlockChain.Chain {
	//	for _, peerBlock := range peerBlockList {
	//		var i int
	//		var myBlock p2.Block
	//		myBlockList := sbc.bc.Get(height)
	//		for i, myBlock = range myBlockList{
	//			if peerBlock.Header.Hash == myBlock.Header.Hash {
	//				break
	//			}
	//		}
	//		if i >= len(myBlockList) {
	//			sbc.bc.Insert(peerBlock)
	//		}
	//	}
	//}
}

/**
Encode SyncBlockChain to json
 */
func(sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	json, err := sbc.bc.EncodeToJson()
	return json, err
}

// GenBlock(): This function generates a new block after the current highest block.
// You may consider it "create the next block".
func(sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie, nonce string) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	curBlockHeight := sbc.bc.Length
	timeStamp := time.Now().Unix()
	//parentHash is the hash of the block at previous height.
	parentHash := sbc.bc.Chain[curBlockHeight][0].Header.Hash
	block := p2.NewBlock(curBlockHeight + 1, timeStamp, parentHash, mpt, nonce)
	sbc.bc.Insert(block)
	return block
}

/**
Show SyncBlockChain
 */
func(sbc *SyncBlockChain) Show() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Show()
}

/**
synchronized version:
GetLatestBlocks(): This function returns the list of blocks of height "BlockChain.length".
 */
func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

/**
synchronized version:
GetParentBlock(): This function takes a block as the parameter, and returns its parent block.
 */
func (sbc *SyncBlockChain) GetParentBlock(b p2.Block) (p2.Block, bool){
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(b)
}