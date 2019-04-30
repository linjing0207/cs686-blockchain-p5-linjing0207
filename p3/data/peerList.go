package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

type PeerList struct {
	selfId int32
	peerMap map[string]int32  //PeerMap maps IP Address to its ID.
	maxLength int32
	mux sync.Mutex
}

/**
NewPeerList() is the initial function of PeerList structure.
 */
func NewPeerList(id int32, maxLength int32) PeerList {
	peerList := PeerList{selfId:id, peerMap:make(map[string]int32), maxLength:maxLength}
	return peerList
}

/**
PeerMap maps IP Address to its ID.
 */
func(peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peers.peerMap[addr] = id
}

/**
Delete IP Address in PeerMap
 */
func(peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	delete(peers.peerMap, addr)
}

/**
Every user would hold a PeerList of up to 32 peer nodes. (32 is the number Ethereum uses.)
The PeerList can temporarily hold more than 32 nodes, but before sending HeartBeats,
a node will first re-balance the PeerList by choosing the 32 closest peers.
"Closest peers" is defined by this: Sort all peers' Id, insert SelfId,
consider the list as a cycle,and choose 16 nodes at each side of SelfId.
For example, if SelfId is 10, PeerList is [7, 8, 9, 15, 16],
then the closest 4 nodes are [8, 9, 15, 16].
 */
func(peers *PeerList) Rebalance() {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	//sort
	//choose 16 nodes at each side of SelfId = delete redundant nodes
	if len(peers.peerMap) > int(peers.maxLength) {
		//var deleteList []PairList
		pl := make(PairList, len(peers.peerMap))
		i := 0
		for k, v := range peers.peerMap {
			pl[i] = Pair{k, int(v)}
			i++
		}
		sort.Sort(pl)
		newList := make(PairList, 0)
		newPeerMap := make(map[string]int32)
		for i := range pl {
			if int32(pl[i].Value) > peers.selfId {
				var nodes = int(peers.maxLength)/2
				if i-nodes < 0 {
					newList = append(newList, pl[:i]...)
					newList = append(newList, pl[(len(pl)-nodes+1):]...)
					newList = append(newList, pl[i:i+nodes]...)
				} else if i+nodes > len(pl) {
					newList = append(newList, pl[i-nodes:i]...)
					newList = append(newList, pl[i:]...)
					newList = append(newList, pl[:(nodes-(len(pl)-i))]...)
				} else {
					newList = append(pl[i-nodes:i], pl[i:i+nodes]...)
				}
				break
			}
		}
		for _, v := range newList {
			newPeerMap[v.Key] = int32(v.Value)
		}
		peers.peerMap = newPeerMap
	}
}

//for sorting map by value(id)
type Pair struct {
	Key string
	Value int
}

type PairList []Pair


func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }


/**
For example, it returns "This is PeerMap: \n addr=127.0.0.1, id=1".
 */
func(peers *PeerList) Show() string {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	str := "This is PeerMap: \n"
	for addr,id := range peers.peerMap {
		str += "addr=" + addr + ", id=" + strconv.Itoa(int(id)) + "\n"
	}
	return str
}

/**
Register() is used to set ID. You can consider it as "SetId()".
 */
func(peers *PeerList) Register(id int32) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

/**
return peerMap from peerList
 */
func(peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peerMap := peers.peerMap
	return peerMap
}

/**
Get current selfId from peerList
 */
func(peers *PeerList) GetSelfId() int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.selfId
}

/**
Encode peerMap to json
 */
func(peers *PeerList) PeerMapToJson() (string, error) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	jsonPeerMap, err := json.Marshal(peers.peerMap)
	return string(jsonPeerMap), err
}


/**
InjectPeerMapJson() inserts every entries(every <addr, id> pair) of the parameter "peerMapJsonStr" into your own PeerMap,
except the entry whose address is your own local address.
 */
func(peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peerMap := make(map[string]int32)
	err := json.Unmarshal([]byte(peerMapJsonStr), &peerMap)
	if err != nil {
		fmt.Println("Umarshal failed:", err)
	}
	for addr, id := range peerMap {
		if addr != selfAddr {
			_, ok := peers.peerMap[addr]
			if ok == false { // not exist
				peers.peerMap[addr] = id
			}
		}
	}
}


func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}