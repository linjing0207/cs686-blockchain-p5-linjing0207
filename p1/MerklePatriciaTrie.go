package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strconv"
	"strings"
)

/**
This class used for Extension and Leaf node
 */
type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

/**
Enum Node
This class represent a node of type Branch, Leaf, Extension, or Null.
 */
type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

/**
Struct MerklePatriciaTrie
This class represent a Merkle Patricia Trie. It has two variables: "db" and "root".
Variable "db" is a HashMap. The key of the HashMap is a Node's hash value. The value of the HashMap is the Node.
Variable "root" is a String, which is the hash value of the root node.
 */
type MerklePatriciaTrie struct {
	db   map[string]Node //map: key( Node's hash value) value(Node)
	root string
}
/**
store key and value
Example:
"a" -> "apple"
 */

var mptMap = make(map[string]string)


/**
Get the root from mpt.
Return: the root of mpt (string).
 */
func (mpt *MerklePatriciaTrie) GetRoot() string{
	return mpt.root
}
/**
Description:
The Get function takes a key as argument,
traverses down the Merkle Patricia Trie to find the value, and returns it.
If the key doesn't exist, it will return an empty string.
(for the Go version: if the key is nil, Get returns an empty string.)
Arguments: key (string) --"abc"
Return: the value stored for that key (string).
Rust function definition: fn get(&mut self, key: &str) -> String
Go function definition: func (mpt *MerklePatriciaTrie) Get(key string) string
 */
func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	var value string
	var err error
	if mpt == nil {
		value = ""
		err = errors.New("path_not_found")
	} else {
		//convert string to hex array [1,6,1]
		hex_array := stringToHex_array(key)
		if key == "" {
			value = ""
			err = errors.New("path_not_found")
		} else {
			//get helper
			value, err = mpt.get_helper(hex_array, mpt.root)
		}
	}
	return value, err
}

/**
Description:
Insert() function takes a pair of <key, value> as arguments.
It will traverse down the Merkle Patricia Trie, find the right place to insert the value,and do the insertion.
Arguments: key(String), value(String)
For example: ("a","apple")
Return: None
 */
func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	hex_array := stringToHex_array(key)
	node_hash := mpt.insert_helper(hex_array, new_value, mpt.root)
	mpt.root = node_hash
}

/**
Description:
The Delete function takes a key as argument,
traverses the Merkle Patricia Trie and finds that key.
If the key exists, delete the corresponding value and re-balance the trie if necessary,
then return an empty string; if the key doesn't exist, return "path_not_found".
Arguments: key (string)
Return: string
Rust function definition: fn delete(&mut self, key: &str) -> String
Go function definition: func (mpt *MerklePatriciaTrie) Delete(key string) string
 */
func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	var value string
	var err error
	hex_array := stringToHex_array(key)
	if mpt == nil {
		value = ""
		err = errors.New("path_not_found")
	} else {
		//var result string
		value = mpt.delete_helper(hex_array, mpt.root)
		if value == "path_not_found" {
			value = ""
			err = errors.New("path_not_found")
		} else {
			mpt.root = value
		}

	}
	return value, err
}

/**
Description:
The compact_encode function takes an array of numbers as input
(each number is between 0 and 15 included, representing a single hex digit),
and returns an array of numbers according to the compact encoding rules in the github wiki page
under "Compact encoding of hex sequence with optional terminator").
Each number in the output is between 0 and 255 included
(representing an ASCII-encoded letter, or for the first value it represents the node type as per the wiki page).
You may find a Python version in this Link (Links to an external site.)
Links to an external site., but be mindful that the return type is different!
Arguments: hex_array(array of u8)
Return: array of u8
Rust function definition: compact_encode(hex_array: Vec<u8>) -> Vec<u8>
Example: input=[1, 6, 1], encoded_array=[1, 1, 6, 1], output=[17, 97]
 */
func compact_encode(hex_array []uint8) []uint8 {
	//[1, 6, 1]
	//add the flags method
	var encoded_array []uint8
	var term int
	var output []uint8
	if len(hex_array) == 0 {

	} else {
		//if is Leaf Node
		if hex_array[len(hex_array)-1] == 16 { //[]uint8{0, 15, 1, 12, 11, 8, 16}
			term = 1
			//delete the last element:16
			hex_array = hex_array[:len(hex_array)-1] //[]uint8{0, 15, 1, 12, 11, 8}
		} else {
			term = 0
		}
		oddlen := len(hex_array) % 2
		flags := uint8(2*term + oddlen)

		//the length is even
		if oddlen == 0 {
			encoded_array = append(encoded_array, flags, 0) // flags = 1 or 3?
		} else {
			encoded_array = append(encoded_array, flags)
		}

		//add flags with original array
		encoded_array = append(encoded_array, hex_array...)
		for i := 0; i < len(encoded_array); i += 2 {
			temp := encoded_array[i]*16 + encoded_array[i+1]
			//uint := stringToUint8(temp)
			output = append(output, temp)
		}
	}
	//output1 := uint8(output)
	return output //[0, 97]
}

//func main()  {
//	hex_array := []uint8{6,1}
//	fmt.Println(compact_decode(compact_encode(hex_array)))
//}

/**
Description:
This function reverses the compact_encode() function.
Arguments: hex_array(array of u8)
Return: array of u8
Rust function definition: compact_decode(encoded_arr: Vec<u8>) -> Vec<u8>
Example: input=[17, 97], output=[1, 6, 1]
 */
// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	var decode_array []uint8
	//1.traverse encoded_arr, convert to HEX value
	for _, v := range encoded_arr { //[0, 97] or [17, 97]
		//0/16,0%16
		decode_array = append(decode_array, v/16, v%16) // [0, 0, 6, 1]
	}

	if len(decode_array) == 0 {
		decode_array = []uint8{}
	} else {
		//2.remove the flags
		if decode_array[0]%2 == 0 {
			decode_array = decode_array[2:] //[0, 0, 6, 1]
		} else {
			decode_array = decode_array[1:] //[0, 1, 6, 1]
		}
	}
	return decode_array
}

/**
Description:
The get_helper function helps Get function to find value.
Arguments: hex_array(array of u8)
Return: the value stored for that key (string), error.
 */
func (mpt *MerklePatriciaTrie) get_helper(hex_array []uint8, hash string) (string, error) {
	var value string
	var err error
	curNode := mpt.db[hash]

	switch curNode.node_type {
	case 0: //NULL
		value = ""
		err = errors.New("path_not_found")
	case 1: //Branch Node
		value, err = mpt.branch_get_helper(hex_array, hash)
	case 2:                                                //Ext or Leaf
		encodedPrefix := curNode.flag_value.encoded_prefix //[17,97]
		decode_array := compact_decode(encodedPrefix)      //[1,6,1]
		prefix := encodedPrefix[0] / 16
		//check every element in two arrays(hex_array and decode_array)
		if len(hex_array) < len(decode_array) {
			value = ""
			err = errors.New("path_not_found")
		} else {
			for i := 0; i < len(decode_array); i++ {
				if hex_array[i] != decode_array[i] {
					value = ""
					err = errors.New("path_not_found")
				}
			}
			hex_array = hex_array[len(decode_array):]
			if prefix == 2 || prefix == 3 { //leaf
				value, err = mpt.leaf_get_helper(hex_array, hash)
			} else { //Ext
				value, err = mpt.extension_get_helper(hex_array, hash)
			}
		}

	}
	return value, err
}

/**
Description:
The branch_get_helper function helps to find value for current node is branch node.
Arguments: hex_array(array of u8)
Return: the value stored for that key (string), error.
 */
func (mpt *MerklePatriciaTrie) branch_get_helper(hex_array []uint8, hash string) (string, error) {
	var value string
	var err error
	curNode := mpt.db[hash]

	nextNode := curNode.branch_value[hex_array[0]]
	rest_path := hex_array[1:]
	//if the hash value of next node is nil, Get returns an empty string.
	if nextNode == "" {
		value = ""
		err = errors.New("path_not_found")
	} else {
		value, err = mpt.get_helper(rest_path, nextNode)
	}
	return value, err
}

/**
Description:
The leaf_get_helper function helps to find value for current node is leaf node.
Arguments: hex_array(array of u8)
Return: the value stored for that key (string), error.
 */
func (mpt *MerklePatriciaTrie) leaf_get_helper(hex_array []uint8, hash string) (string, error) {
	var value string
	var err error
	curNode := mpt.db[hash]
	if len(hex_array) == 0 {
		value = curNode.flag_value.value
	} else {
		value = ""
		err = errors.New("path_not_found")
	}
	return value, err
}

/**
Description:
The extension_get_helper function helps to find value for current node is extension node.
Arguments: hex_array(array of u8)
Return: the value stored for that key (string), error.
 */
func (mpt *MerklePatriciaTrie) extension_get_helper(hex_array []uint8, hash string) (string, error) {
	var value string
	var err error
	curNode := mpt.db[hash]
	if len(hex_array) == 0 {
		value = mpt.db[curNode.flag_value.value].branch_value[16]
	} else {
		value, err = mpt.get_helper(hex_array, curNode.flag_value.value)
	}
	return value, err
}

/**
Description:
The insert_helper function helps Insert function.
Arguments: hex_array(array of u8), new_value(string), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) insert_helper(hex_array []uint8, new_value string, hash string) string {
	curNode := mpt.db[hash]
	//delete(mpt.db, hash)
	var node_hash string //hash value of current node

	switch curNode.node_type {
	case 0: //NULL
		delete(mpt.db, hash)
		//insert root node, it should be leaf node
		node_hash = mpt.create_leaf_node(hex_array, new_value)
		//mpt.db = curNode
	case 1: //Branch Node
		//delete(mpt.db, hash)
		node_hash = mpt.branch_insert_helper(hex_array, new_value, hash)
	case 2:                                                //Ext or Leaf
		encodedPrefix := curNode.flag_value.encoded_prefix //[17,97]
		prefix := encodedPrefix[0] / 16
		//current node is leaf node, if we create a new leaf node with the original value, we keep it.
		//if we create a new leaf node with the new value, we delete the old one which store in db
		if prefix == 2 || prefix == 3 { //Leaf Node
			node_hash = mpt.leaf_insert_helper(hex_array, new_value, hash)
		} else { //Ext Node
			node_hash = mpt.extension_insert_helper(hex_array, new_value, hash)
		}
	}
	return node_hash
}

/**
Description:
The branch_insert_helper function helps to insert new elements for current node is branch node.
Arguments: hex_array(array of u8), new_value(string), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) branch_insert_helper(hex_array []uint8, new_value string, hash string) string {
	curNode := mpt.db[hash]
	delete(mpt.db, hash)
	var node_hash string //hash value of current node

	if len(hex_array) == 0 { //insert value
		curNode.branch_value[16] = new_value
	} else {
		//branch-2
		next_node_hash := curNode.branch_value[hex_array[0]]
		if next_node_hash == "" { //create new leaf node
			leaf := mpt.create_leaf_node(hex_array[1:], new_value)
			//update current node
			curNode.branch_value[hex_array[0]] = leaf
		} else { //2.2find next
			update_hash := mpt.insert_helper(hex_array[1:], new_value, next_node_hash)
			//update current node
			curNode.branch_value[hex_array[0]] = update_hash
		}
	}
	//store in the map
	mpt.db[curNode.hash_node()] = curNode
	//return the hash value of current node
	node_hash = curNode.hash_node()
	return node_hash
}

/**
Description:
The extension_insert_helper function helps to insert new elements for current node is extension node.
Arguments: hex_array(array of u8), new_value(string), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) extension_insert_helper(hex_array []uint8, new_value string, hash string) string {
	curNode := mpt.db[hash]
	var node_hash string                               //hash value of current node
	encodedPrefix := curNode.flag_value.encoded_prefix //[17,97]
	decode_array := compact_decode(encodedPrefix)      //[1,6,1]
	commonLen := common_length(hex_array, decode_array)
	//prefix := encodedPrefix[0]/16
	if len(hex_array) == 0 { // 1
		if len(decode_array) == 1 {
			//create branch
			branch_value := [17]string{}
			branch_value[decode_array[0]] = curNode.flag_value.value
			branch_value[16] = new_value
			node_hash = mpt.create_branch_node(branch_value)
		} else { //len(decode_array) > 1
			//create extension
			ext := mpt.create_extension_node(decode_array[1:], curNode.flag_value.value)
			//create branch
			branch_value := [17]string{}
			branch_value[decode_array[0]] = ext
			branch_value[16] = new_value
			node_hash = mpt.create_branch_node(branch_value)
		}
	} else if commonLen == 0 { //2. hex_array=[6,2] decode_array=[7,8,9]
		if len(decode_array) == 1 {
			//create leaf
			leaf := mpt.create_leaf_node(hex_array[1:], new_value)
			//create branch
			branch_value := [17]string{}
			branch_value[decode_array[0]] = curNode.flag_value.value
			branch_value[hex_array[0]] = leaf
			node_hash = mpt.create_branch_node(branch_value)
		} else { //len(decode_array)>1
			//create leaf
			leaf := mpt.create_leaf_node(hex_array[1:], new_value)
			//create ext
			ext := mpt.create_extension_node(decode_array[1:], curNode.flag_value.value)
			//create branch
			branch_value := [17]string{}
			branch_value[hex_array[0]] = leaf
			branch_value[decode_array[0]] = ext
			node_hash = mpt.create_branch_node(branch_value)
		}
	} else if commonLen == len(decode_array) { //3. hex_array=[6,2,3] decode_array=[6,2]
		//find next node
		update_hash := mpt.insert_helper(hex_array[commonLen:], new_value, curNode.flag_value.value)
		curNode.flag_value.value = update_hash
		mpt.db[curNode.hash_node()] = curNode
		node_hash = curNode.hash_node()
	} else { //4.commonLen < len(decode_array)
		if commonLen == len(hex_array) { //4.1
			if commonLen < len(decode_array)-1 { //4.1.1
				//create extension2
				ext2 := mpt.create_extension_node(decode_array[commonLen+1:], curNode.flag_value.value)
				//create branch
				branch_value := [17]string{}
				branch_value[decode_array[commonLen]] = ext2
				branch_value[16] = new_value
				branch := mpt.create_branch_node(branch_value)
				//create ext1
				node_hash = mpt.create_extension_node(decode_array[:commonLen], branch)
			} else { //4.1.2 commonLen = len(decode_array) - 1
				//create branch
				branch_value := [17]string{}
				branch_value[decode_array[0]] = curNode.flag_value.value
				branch_value[16] = new_value
				branch := mpt.create_branch_node(branch_value)
				//create ext
				node_hash = mpt.create_extension_node(decode_array[:commonLen], branch)
			}
		} else { //commonLen < len(hex_array)
			if commonLen < len(decode_array)-1 {
				//create leaf
				leaf := mpt.create_leaf_node(hex_array[commonLen+1:], new_value)
				//create ext2
				ext2 := mpt.create_extension_node(decode_array[commonLen+1:], curNode.flag_value.value)
				//create branch
				branch_value := [17]string{}
				branch_value[decode_array[commonLen]] = ext2
				branch_value[hex_array[commonLen]] = leaf
				branch := mpt.create_branch_node(branch_value)
				//create ext1
				node_hash = mpt.create_extension_node(hex_array[:commonLen], branch)

			} else { //commonLen = len(decode_array) - 1
				//create leaf
				leaf := mpt.create_leaf_node(hex_array[commonLen+1:], new_value)
				//create branch
				branch_value := [17]string{}
				branch_value[hex_array[commonLen]] = leaf
				branch_value[decode_array[commonLen]] = curNode.flag_value.value
				branch := mpt.create_branch_node(branch_value)
				//create ext
				node_hash = mpt.create_extension_node(hex_array[:commonLen], branch)
			}
		}
	}
	return node_hash
}

/**
Description:
The leaf_insert_helper function helps to insert new elements for current node is leaf node.
Arguments: hex_array(array of u8), new_value(string), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) leaf_insert_helper(hex_array []uint8, new_value string, hash string) string {
	curNode := mpt.db[hash]
	delete(mpt.db, hash)
	var node_hash string                               //hash value of current node
	encodedPrefix := curNode.flag_value.encoded_prefix //[17,97]
	decode_array := compact_decode(encodedPrefix)      //[1,6,1]
	commonLen := common_length(hex_array, decode_array)
	//prefix := encodedPrefix[0]/16
	if commonLen == len(hex_array) && commonLen == len(decode_array) { //1. totally match
		curNode.flag_value.value = new_value //update value
		mpt.db[curNode.hash_node()] = curNode
		node_hash = curNode.hash_node()
		//since we update the value of the node, delete the old one
	} else if commonLen == 0 { //2. totally un match
		if len(decode_array) == 0 {
			//create leaf
			leaf := mpt.create_leaf_node(hex_array[1:], new_value)
			//create branch
			branch_value := [17]string{}
			branch_value[hex_array[0]] = leaf
			branch_value[16] = curNode.flag_value.value
			node_hash = mpt.create_branch_node(branch_value)
		} else if len(hex_array) == 0 {
			//create leaf
			leaf := mpt.create_leaf_node(decode_array[1:], curNode.flag_value.value)
			//create branch
			branch_value := [17]string{}
			branch_value[decode_array[0]] = leaf
			branch_value[16] = new_value
			node_hash = mpt.create_branch_node(branch_value)
			//since the new leaf node has same value, the hash value store in db will be same.
			//we don't have to delete the old one
		} else { //2.3hex=[1,2] cur=[3,4]
			//create leaf1
			leaf1 := mpt.create_leaf_node(hex_array[1:], new_value)
			//create leaf2
			leaf2 := mpt.create_leaf_node(decode_array[1:], curNode.flag_value.value)
			branch_value := [17]string{}
			branch_value[hex_array[0]] = leaf1
			branch_value[decode_array[0]] = leaf2
			node_hash = mpt.create_branch_node(branch_value)
		}

	} else if commonLen == len(decode_array) { //3
		//create leaf node
		leaf := mpt.create_leaf_node(hex_array[commonLen+1:], new_value)
		//create branch node
		branch_value := [17]string{}
		branch_value[hex_array[commonLen]] = leaf
		branch_value[16] = curNode.flag_value.value
		branch := mpt.create_branch_node(branch_value)
		//create extension node
		node_hash = mpt.create_extension_node(hex_array[:commonLen], branch)

	} else if commonLen == len(hex_array) { //4
		//create leaf
		leaf := mpt.create_leaf_node(decode_array[commonLen+2:], curNode.flag_value.value)
		branch_value := [17]string{}
		//create leaf node
		branch_value[decode_array[commonLen]] = leaf
		branch_value[16] = new_value
		//create branch
		branch := mpt.create_branch_node(branch_value)
		//create extension
		node_hash = mpt.create_extension_node(hex_array[:commonLen-1], branch)
	} else if commonLen <= len(hex_array) && commonLen <= len(decode_array) { //5
		//ok
		//create leaf1
		leaf1 := mpt.create_leaf_node(hex_array[commonLen+1:], new_value)
		//create leaf2
		leaf2 := mpt.create_leaf_node(decode_array[commonLen+1:], curNode.flag_value.value)
		//create branch
		branch_value := [17]string{}
		branch_value[hex_array[commonLen]] = leaf1
		branch_value[decode_array[commonLen]] = leaf2
		branch := mpt.create_branch_node(branch_value)
		//create extension
		node_hash = mpt.create_extension_node(hex_array[:commonLen], branch)
	}
	return node_hash
}

/**
Description:
Create a leaf node.
Arguments: hex_array(array of u8), new_value(string))
return:hash value of leaf node (string)
 */
func (mpt *MerklePatriciaTrie) create_leaf_node(hex_array []uint8, new_value string) string {
	//add 16 to the end of array
	hex_array = append(hex_array, 16)
	flag_value := Flag_value{compact_encode(hex_array), new_value}
	leaf_node := Node{2, [17]string{}, flag_value}
	mpt.db[leaf_node.hash_node()] = leaf_node
	return leaf_node.hash_node()
}

/**
Description:
Create a branch node.
Arguments: branch_value([17]string)
return:hash value of branch node (string)
 */
func (mpt *MerklePatriciaTrie) create_branch_node(branch_value [17]string) string {

	branch_node := Node{1, branch_value, Flag_value{}}
	mpt.db[branch_node.hash_node()] = branch_node
	return branch_node.hash_node()
}

/**
Description:
Create a extension node.
Arguments: hex_array(array of u8), hash_value(string))
return:hash value of extension node (string)
 */
func (mpt *MerklePatriciaTrie) create_extension_node(hex_array []uint8, hash_value string) string {
	flag_value := Flag_value{compact_encode(hex_array), hash_value}
	extention_node := Node{2, [17]string{}, flag_value}
	mpt.db[extention_node.hash_node()] = extention_node
	return extention_node.hash_node()
}

/**
Description:
The delete_helper function helps Delete function.
Arguments: hex_array(array of u8), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) delete_helper(hex_array []uint8, hash string) string {
	//var value string
	curNode := mpt.db[hash]
	//delete(mpt.db, hash)
	var node_hash string //hash value of current node
	switch curNode.node_type {
	case 0: //NULL
		node_hash = ""
	case 1: //Branch
		node_hash = mpt.branch_delete_helper(hex_array, hash)
	case 2:                                                //Ext or Leaf
		encodedPrefix := curNode.flag_value.encoded_prefix //[17,97]
		prefix := encodedPrefix[0] / 16
		if prefix == 2 || prefix == 3 { //Leaf Node
			node_hash = mpt.leaf_delete_helper(hex_array, hash)
		} else { // Ext
			node_hash = mpt.extension_delete_helper(hex_array, hash)
		}
	}
	return node_hash
}

/**
Description:
The branch_delete_helper function helps to delete the corresponding value for current node is branch node.
Arguments: hex_array(array of u8), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) branch_delete_helper(hex_array []uint8, hash string) string {
	var node_hash string
	curNode := mpt.db[hash]
	if len(hex_array) == 0 { //1
		if curNode.branch_value[16] == "" { //1.1
			node_hash = "path_not_found"
		} else { //1.2
			//delete old branch node
			delete(mpt.db, hash)
			//update value
			curNode.branch_value[16] = ""
			sum := elements_sum(curNode.branch_value)
			if sum == 1 { //1.2.1
				//next_hash := find_next_node(curNode.branch_value)
				index := find_next_node(curNode.branch_value)
				next_hash := curNode.branch_value[index]
				next_node := mpt.db[next_hash]
				if next_node.node_type == 1 { //1.2.1.3 Branch
					//get new hex_array
					arr := []uint8{uint8(index)}
					//create ext
					node_hash = mpt.create_extension_node(arr, next_hash)
				} else { //next_node.node_type == 2
					prefix := next_node.flag_value.encoded_prefix[0] / 16
					if prefix == 2 || prefix == 3 { //1.2.1.1 Leaf Node
						//delete leaf
						delete(mpt.db, next_hash)
						//combine arr
						arr := []uint8{uint8(index)}
						leaf_remain := compact_decode(next_node.flag_value.encoded_prefix)
						arr = append(arr, leaf_remain...)
						//create leaf
						node_hash = mpt.create_leaf_node(arr, next_node.flag_value.value)
					} else { //1.2.1.2 Ext
						//delete ext
						delete(mpt.db, next_hash)
						//combine arr
						arr := []uint8{uint8(index)}
						ext_remain := compact_decode(next_node.flag_value.encoded_prefix)
						arr = append(arr, ext_remain...)
						//create ext
						node_hash = mpt.create_extension_node(arr, next_node.flag_value.value)
					}
				}
			} else { //1.2.2 sum>1
				mpt.db[curNode.hash_node()] = curNode
				node_hash = curNode.hash_node()
			}
		}
	} else { //2.len(hex) !=0
		if curNode.branch_value[hex_array[0]] == "" { //2.1
			node_hash = "path_not_found"
		} else { //2.2
			return_value := mpt.delete_helper(hex_array[1:], curNode.branch_value[hex_array[0]])
			if return_value == "path_not_found" { //2.2.1
				node_hash = "path_not_found"
			} else if return_value == "" { //2.2.2 already delete the next node, and nothing left
				//remove branch node
				delete(mpt.db, hash)
				curNode.branch_value[hex_array[0]] = ""
				if elements_sum(curNode.branch_value) > 1 { //2.2.2.1
					mpt.db[curNode.hash_node()] = curNode
					node_hash = curNode.hash_node()
				} else { //2.2.2.2 elements_sum(curNode.branch_value) = 1
					if curNode.branch_value[16] != "" { //2.2.2.2.1 not the value
						//create leaf
						node_hash = mpt.create_leaf_node([]uint8{}, curNode.branch_value[16])
					} else { //2.2.2.2.2 b_v[0~15]
						index := find_next_node(curNode.branch_value)
						next_hash := curNode.branch_value[index]
						next_node := mpt.db[next_hash]

						if next_node.node_type == 1 { //Branch
							//get hex_array
							arr := []uint8{uint8(index)}
							//create ext
							node_hash = mpt.create_extension_node(arr, next_hash)
						} else { //next_node.node_type == 2
							prefix := next_node.flag_value.encoded_prefix[0] / 16
							if prefix == 2 || prefix == 3 { //Leaf Node
								//remove leaf
								delete(mpt.db, next_hash)
								//combine arr
								arr := []uint8{uint8(index)}
								leaf_remain := compact_decode(next_node.flag_value.encoded_prefix)
								arr = append(arr, leaf_remain...)
								//create leaf
								node_hash = mpt.create_leaf_node(arr, next_node.flag_value.value)
							} else { // Ext
								//delete ext
								delete(mpt.db, next_hash)
								//combine arr
								arr := []uint8{uint8(index)}
								ext_remain := compact_decode(next_node.flag_value.encoded_prefix)
								arr = append(arr, ext_remain...)
								//create ext
								node_hash = mpt.create_extension_node(arr, next_node.flag_value.value)
							}
						}
					}
				}
			} else { //2.2.3 hash
				delete(mpt.db, hash)
				curNode.branch_value[hex_array[0]] = return_value
				mpt.db[curNode.hash_node()] = curNode
				node_hash = curNode.hash_node()
			}
		}
	}

	return node_hash
}

/**
Description:
The extension_delete_helper function helps to delete the corresponding value for current node is extension node.
Arguments: hex_array(array of u8), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) extension_delete_helper(hex_array []uint8, hash string) string {
	var node_hash string
	curNode := mpt.db[hash]
	decode_array := compact_decode(curNode.flag_value.encoded_prefix) //[1,6,1]
	commonLen := common_length(hex_array, decode_array)
	if commonLen != len(decode_array) { //1
		node_hash = "path_not_found"
	} else { //2
		//the next node hash of ext node is curNode.flag_value.value
		retrun_value := mpt.delete_helper(hex_array[commonLen:], curNode.flag_value.value)
		if retrun_value == "path_not_found" { //2.1
			node_hash = "path_not_found"
		} else { //2.3 retrun_value == hash
			return_node := mpt.db[retrun_value]

			if return_node.node_type == 1 { //Branch
				//delete old ext
				delete(mpt.db, hash)
				//update value
				curNode.flag_value.value = retrun_value
				//store in db
				mpt.db[curNode.hash_node()] = curNode
				node_hash = curNode.hash_node()
			} else { //Ext or Leaf
				prefix := return_node.flag_value.encoded_prefix[0] / 16
				if prefix == 2 || prefix == 3 { //Leaf Node
					//remove old ext, leaf
					delete(mpt.db, hash)
					delete(mpt.db, retrun_value)
					//get remains
					ext_remain := compact_decode(curNode.flag_value.encoded_prefix)
					leaf_remain := compact_decode(return_node.flag_value.encoded_prefix)
					//combine
					var arr []uint8
					arr = append(arr, ext_remain...)
					arr = append(arr, leaf_remain...)
					//create leaf
					node_hash = mpt.create_leaf_node(arr, return_node.flag_value.value)
				} else { // Ext
					//remove ext1, ext2
					delete(mpt.db, hash)
					delete(mpt.db, retrun_value)
					//get remains
					ext1_remain := compact_decode(curNode.flag_value.encoded_prefix)
					ext2_remain := compact_decode(return_node.flag_value.encoded_prefix)
					//combine
					var arr []uint8
					arr = append(arr, ext1_remain...)
					arr = append(arr, ext2_remain...)
					//create new ext
					node_hash = mpt.create_extension_node(arr, return_node.flag_value.value)
				}
			}
		}
		// which will not happen, since the next node ext node is branch, will not be totally delete
		//else if retrun_value == "" { //2.2}

	}

	return node_hash
}

/**
Description:
The leaf_delete_helper function helps to delete the corresponding value for current node is leaf node.
Arguments: hex_array(array of u8), hash(string)
Return: the value stored for that key (string)
 */
func (mpt *MerklePatriciaTrie) leaf_delete_helper(hex_array []uint8, hash string) string {
	var value string
	curNode := mpt.db[hash]
	decode_array := compact_decode(curNode.flag_value.encoded_prefix) //[1,6,1]
	commonLen := common_length(hex_array, decode_array)
	if commonLen == len(hex_array) && commonLen == len(decode_array) { //1
		//delete leaf node
		delete(mpt.db, hash)
		value = ""
	} else { //2 hex != cur
		value = "path_not_found"
	}
	return value
}

/**
Description:
This function converts string to array of unit8.
Arguments: s(string)
Return: hex_array(array of u8)
Example: string="abc", hex_array=[6, 1, 6, 2, 6, 3]
 */
func stringToHex_array(s string) []uint8 {
	src := []uint8(s)                          //[97 98 99]
	encodedStr := hex.EncodeToString(src)      //-> 616263
	str_array := strings.Split(encodedStr, "") //-> [6,1,6,2,6,3]
	//convert string to uint8
	var hex_array []uint8
	for _, v := range str_array {
		hex_array = append(hex_array, stringToUint8(v))
	}
	//-> [6,1,6,2,6,3]
	return hex_array
}

/**
Description:
This function converts type string to type unit8.
Arguments: s(string)
Return: uint8
 */
func stringToUint8(s string) uint8 {
	var f uint64
	f, _ = strconv.ParseUint(s, 16, 64)
	return uint8(f)
}

/**
Description:
This function converts array of unit8 to string.
Arguments: hex_array(array of u8)
Return: s(string)
Example: hex_array=[6,1,6,2,6,3], string="abc"
 */
func Hex_arrayToString(hex_array []uint8) string {
	//key := Hex_arrayToString(hex_array)
	dec_array := compact_encode(hex_array) //[0,97]
	//???
	if dec_array[0] == 0 {
		dec_array = dec_array[1:]
	}
	//var str string
	//for _,v := range hex_array{
	//	str += string(strconv.Itoa(int(v))) //616263
	//}
	//fmt.Println(str)
	//dec_array, err := hex.DecodeString(str) //[97 98 99]
	//if err != nil {
	//	panic(err)
	//}
	//
	return string(dec_array)
}

/**
Description:
This function counts the common length of two arrays.
Arguments: array1 ([]uint8), array2 ([]uint8)
Return: int
 */
func common_length(array1 []uint8, array2 []uint8) int {
	comLen := len(array1)
	if len(array1) >= len(array2) {
		comLen = len(array2)
	}
	i := 0
	for ; i < comLen; i++ {
		if array1[i] != array2[i] {
			break
		}
	}
	return i
}

/**
Description:
This function calculates the number of elements in a array
Arguments: branch_value([17]string)
Return: int
 */
func elements_sum(branch_value [17]string) int {
	var sum int
	for _, v := range branch_value {
		if v != "" {
			sum++
		}
	}
	return sum
}

/**
Description:
This function returns the not null index from a array
Arguments: branch_value([17]string)
Return: int
 */
func find_next_node(branch_value [17]string) int {
	var index int
	for i, v := range branch_value {
		if v != "" {
			index = i
		}
	}
	return index
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

/**
Description: This function takes a node as the input, hash the node and return the hashed string.
 */
func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			//connect all the elements in the array
			str += v
		}
	case 2:
		//leaf node: rlp value
		//ext node: hash value of next node
		//node.flag_value.prefix
		str = node.flag_value.value
	}
	//encryption
	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

/**
The size is the length of the byte array of the block value.
You have a mpt as the block's value, you convert mpt to byte array, then size equals to the length of that byte array.
 */
func (mpt *MerklePatriciaTrie) MptToByteArray() []byte {
	//1.
	//l := unsafe.Sizeof(mpt)
	//pb := (*[1024]byte)(unsafe.Pointer(&mpt))
	//byteArray := (*pb)[:l]

	//2.
	byteArray := []byte(fmt.Sprintf("%v", mpt))
	return byteArray
}

/**
Description:
This function traverse the mpt and return a map of key and values
Arguments: hash(string), previous（[]uint8）
Return: map[string]string
 */
func (mpt *MerklePatriciaTrie) tranverseMpt(hash string, previous []uint8) map[string]string {

	node := mpt.db[hash]
	switch node.node_type {
	case 1: //branch
		for i, v := range node.branch_value {
			if v != "" {
				mpt.tranverseMpt(v, append(previous, uint8(i)))
			}
		}
	case 2: //leaf or ext
		encodedPrefix := node.flag_value.encoded_prefix //[17,97]
		prefix := encodedPrefix[0] / 16
		if prefix == 2 || prefix == 3 { //Leaf Node
			//previous is hex_array, so have to decode(convert to hex) first
			hex_array := append(previous, compact_decode(node.flag_value.encoded_prefix)...)

			key := Hex_arrayToString(hex_array)
			value := node.flag_value.value
			//store key and value
			mptMap[key] = value
		} else { // Ext
			mpt.tranverseMpt(node.flag_value.value, append(previous, compact_decode(node.flag_value.encoded_prefix)...))
		}
	}
	return  mptMap
}

/**
Description:
This function initialize the mptMap and call tranverseMpt method to get all the key and value pairs.
Arguments: hash(string), previous（[]uint8）
Return: map[string]string
 */
func (mpt *MerklePatriciaTrie) GetMptMap(hash string, previous []uint8) map[string]string {
	mptMap = make(map[string]string)
	mpt.tranverseMpt(hash, previous)
	return mptMap
}




