package data

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

//NewHeartBeatData() is a normal initial function which creates an instance.
func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	//chose hops 3
	hops := int32(3)
	heartBeatData := HeartBeatData{ifNewBlock, id, blockJson, peerMapJson, addr, hops}

	return heartBeatData
}

/**
PrepareHeartBeatData() is used when you want to send a HeartBeat to other peers.
PrepareHeartBeatData would first create a new instance of HeartBeatData,
then decide whether or not you will create a new block and send the new block to other peers.
PrepareHeartBeatData():
Randomly decide if you will generate the next block.
If not, return an HeartBeatData without new block;
if yes, do (1) Randomly create an MPT. (2) Generate the next block. (3) Create a HeartBeatData, add that new block, and return.
 */
func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string, blockJson string) HeartBeatData {
	var ifNewBlock = false
	if blockJson != "" {
		ifNewBlock = true
	}
	//you can set it to an empty string.
	return NewHeartBeatData(ifNewBlock, selfId, blockJson, peerMapJSON, addr)
}