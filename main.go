package main

import (
	"./p3"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//import "./p3/data"

func main() {

	router := p3.NewRouter()
	if len(os.Args) > 1 {
		//var err error
		i, err := strconv.Atoi(os.Args[1])
		if err != nil{
			fmt.Println(err)
			return
		}
		p3.Port = int32(i)
		log.Fatal(http.ListenAndServe(":" + os.Args[1], router))


	} else {
		p3.Port = 6688
		log.Fatal(http.ListenAndServe(":6688", router))
	}
	//data.TestPeerListRebalance()


}

func TestNonce()  {
	rand.Seed(time.Now().Unix())
	var nonce string
	for i := 0; i < 16 ; i++ {
		rand := strconv.FormatInt(int64(rand.Intn(16)), 16)
		nonce += rand
	}


	fmt.Println("nonce:" + nonce + "\n")
	nonceHash := p3.GetNonceHash("", nonce, "")
	fmt.Println("begin:" + nonceHash + "\n")
	for {
		if strings.HasPrefix(nonceHash, "00000") {
			fmt.Println("find")
			break
		} else {
			data, err := strconv.ParseInt(nonce, 16, 64)

			if err != nil {
				fmt.Println(err)
			}
			nonce = strconv.FormatInt(int64(data + 1), 16)
			if len(nonce) > 16 {
				nonce = "0000000000000000"
			}
			nonceHash = p3.GetNonceHash("", nonce, "")
			fmt.Println(nonceHash + "\n")
		}
	}
}

