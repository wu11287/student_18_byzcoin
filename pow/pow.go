package main

import (
	// "hash"
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	crypto "myProject/crypto"
	"os"
	"strconv"
	"sync"
	"time"
)

const MAXINT64 int64 = math.MaxInt64
const m int = 2
const n int = 50

type node struct {
	id        int
	pk        crypto.VrfPubkey
	sk        crypto.VrfPrivkey
	hashRes	  string
}

func newNode(nNode *node, i int) {
	nNode.id = i
	nNode.pk, nNode.sk = crypto.VrfKeygen()
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func main() {
	var wg sync.WaitGroup
	nodes := [n]node{}

	for i := 0; i < n; i++ {
		newNode(&nodes[i], i)
	}

	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	var ips = [n]string{}

	// var ips = [500]string{}
	// for i := 0; i < 250; i++ {
	// 	ips[i] = "192.168.0." + strconv.Itoa(i)
	// }
	// for i := 0; i < 250; i++ {
	// 	ips[i+250] = "192.168.1." + strconv.Itoa(i)
	// }

	for i:=0; i < n; i++ {
		wg.Add(1)
		go hashCompute(ips[i], randomness, nodes[i].pk[:], &nodes[i], &wg)
	}
	for i := 0; i < n; i++ {
		ips[i] = "192.168.0." + strconv.Itoa(i)
	}
	wg.Wait()

	for _, v := range nodes {
		Doshard(&v, v.id)
	}

	in := time.Since(start)
	fmt.Println("time consumed: ", in)
}

func Doshard(nNode *node, id int) {
	tmp := nNode.hashRes[len(nNode.hashRes)-1:]
	var end int
	for _, v := range tmp {
		end += (int(v) - '0') 
	}

	sIdx := end % m 
	filename := fmt.Sprintf("shard%v.txt", sIdx)
	var f *os.File
	var err error
	if checkFileIsExist(filename) {
		f, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.Create(filename)
		if err != nil {
			panic(err)
		}
	}
	_, err = fmt.Fprintf(f, "%v ", id)
	if err != nil {
		panic(err)
	}
}


func hashCompute(ip string, rnd []byte, pk []byte, nNode *node, wg *sync.WaitGroup) {
	defer wg.Done()
	var nonce int64

	for nonce < MAXINT64 {
		nNode.hashRes = HashString(ip + string(rnd) + string(pk) + strconv.FormatInt(nonce, 10))
		if nNode.hashRes[:4] != "0000" {
			nonce++
			continue
		} else {
			break
		}
	}
}


// hash 会产生一个64位的输出字符串
func HashString(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}