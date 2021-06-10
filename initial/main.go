// package testmi
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	// "io"
	"math/rand"
	crypto "myProject/crypto"
	"os"
	"sync"
	"time"
)

const denominator float64 = 18446744073709551615

var lock sync.Mutex

const n int = 50 //表示全网节点数目
const m int = 2 //表示分片的个数

type node struct {
	id        int
	weight    int
	pk        crypto.VrfPubkey
	sk        crypto.VrfPrivkey
	rnd       crypto.VrfOutput
	proof     crypto.VrfProof
	pkList    [n]crypto.VrfPubkey
	proofList [n]crypto.VrfProof
	idList    []int
	shardIdx  int
	shardInNode []int
}

type pkAndId struct {
	id int
	pk crypto.VrfPubkey
}

type ProofAndId struct {
	id    int
	proof crypto.VrfProof
}

type RndAndId struct {
	id  int
	rnd crypto.VrfOutput
	p   float64
	in  bool
}

//生成长度为 n 的随机byte数组
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Sotition(nNode *node, msg []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	proof, ok := nNode.sk.ProveMy(msg)
	if !ok {
		fmt.Println("generate proof error!")
	}

	rnd, ok := proof.Hash()
	if !ok {
		fmt.Println("generate rnd error!")
	}
	nNode.rnd = rnd
	nNode.proof = proof
	lock.Unlock()
}

func newNode(nNode *node, i int) {
	nNode.id = i
	// TODO 初始权重设定 
	nNode.weight = rand.Intn(3)+1 //[1,3]
	nNode.pk, nNode.sk = crypto.VrfKeygen()
}

func broadcastPK(ch []chan *pkAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()

	tmp := &pkAndId{id, nNode.pk}
	go func() {
		for i:=0; i< n; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func broadcastProof(ch []chan *ProofAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	tmp := &ProofAndId{id, nNode.proof}

	go func() {
		for i:=0; i < n; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func broadcastRnd(ch []chan *RndAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()
	var p float64
	for i := 0; i < nNode.weight; i++ {
		p = isMeet(nNode)
		if p > 0.7 {
			continue
		} else {
			break;
		}
	}
	var isin bool

	if p > 0.7 {
		isin = false
	} else {
		isin = true
	}

	tmp := &RndAndId{id, nNode.rnd, p, isin}

	go func() {
		for i:=0; i < n; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func isMeet(nNode *node) float64 {
	bytesBuffer := bytes.NewBuffer(nNode.rnd[:])
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	rnd := float64(x)
	if rnd < 0 {
		rnd += denominator
	}

	p := rnd / denominator
	return p
}

func storePk(ch chan *pkAndId, nNode *[n]node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < n; i++ {
		tmp := <-ch
		nNode[id].pkList[tmp.id] = tmp.pk
	}
}

func storeProof(ch chan *ProofAndId, nNode *[n]node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < n; i++ {
		tmp := <-ch
		nNode[id].proofList[tmp.id] = tmp.proof
	}
}

func storeRnd(ch chan *RndAndId, nNode *[n]node, id int, randomness []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < n; i++ {
		tmp := <-ch
		if tmp.in {
			ok := verifyRnd(nNode[id].pkList[tmp.id], nNode[id].proofList[tmp.id], tmp.rnd, randomness, id)
			if ok {
				nNode[id].idList = append(nNode[id].idList, tmp.id)
			}
		}
	}
}

func verifyRnd(pk crypto.VrfPubkey, proof crypto.VrfProof, output crypto.VrfOutput, msg []byte, id int) bool {
	ok, output2 := pk.VerifyMy(proof, msg)

	if !ok {
		fmt.Println("verified error")
		return false
	}
	return output == output2
}


func Doshard(nNode *node, idx int, wg *sync.WaitGroup) {
	defer wg.Done()

	mod := idx % m 
	for _, v := range nNode.idList {
		if v % m == mod {
			nNode.shardInNode = append(nNode.shardInNode, v)
		}
	}
}

// func count() bool {
func main() {
	var wg sync.WaitGroup
	chsPK := make([]chan *pkAndId, n)
	chsProof := make([]chan *ProofAndId, n)
	chsRnd := make([]chan *RndAndId, n)
	nodes := [n]node{}

	rand.Seed(time.Now().Unix())
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		fmt.Println("generate random byte[] failed!")
	}

	for i := 0; i < n; i++ {
		chsPK[i] = make(chan *pkAndId)
		chsProof[i] = make(chan *ProofAndId)
		chsRnd[i] = make(chan *RndAndId)
		newNode(&nodes[i], i)
	}

	for i := 0; i < n; i++ {
		wg.Add(2)
		go broadcastPK(chsPK, &nodes[i], i, &wg) 
		go storePk(chsPK[i], &nodes, i, &wg)
	}

	wg.Wait()

	start := time.Now()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go Sotition(&nodes[i], randomness, &wg)
	}

	wg.Wait()


	for i := 0; i < n; i++ {
		wg.Add(2)
		go broadcastProof(chsProof, &nodes[i], i, &wg)
		go storeProof(chsProof[i], &nodes, i, &wg)
	}
	wg.Wait()


	for i := 0; i < n; i++ {
		wg.Add(2)
		go broadcastRnd(chsRnd, &nodes[i], i, &wg)
		go storeRnd(chsRnd[i], &nodes, i, randomness, &wg)
	}
	wg.Wait()

	var istrue bool
	size := len(nodes[0].idList)
	// fmt.Println(size)
	
	var cnt int = 1
	for i := 1; i < n; i++ {
		tmp := len(nodes[i].idList)
		if tmp == size {
			cnt++
		} else {
			break
		}
	}
	if cnt == n {
		istrue = true
	}
	fmt.Println(istrue) //每个节点都拥有相同的idList，包括未成功加入的节点 

	if size < 4 {
		fmt.Println("the number is not enough, system cannot start!")
	}

	for _, v := range nodes[0].idList {
		wg.Add(1)
		go Doshard(&nodes[v], v, &wg)
	}
	wg.Wait()

	// 确保分片内部的节点数目 > 4
	for _, v := range nodes[0].idList {
		if len(nodes[v].shardInNode) < 4 {
			fmt.Printf("The shard have not enough nodes!\n")
			return
		} 
	}
	
	for i := 0; i < m; i++ { //分片个数
		filename := fmt.Sprintf("shard%v.txt", i)
		f, err := os.Create(filename) 
		if err != nil {
			panic(err)
		}
		defer f.Close()
		for _, v := range nodes[0].idList {
			if v % m == i { //同属于一个分片
				for _, v1 := range nodes[v].shardInNode { 
					_, err = fmt.Fprintf(f, "%v ", v1)
					if err != nil {
						panic(err)
					}
				}
				break
			} else {
				continue
			}
		}
	}

	interval := time.Now().Sub(start) 
	fmt.Printf("time consumed: %v\n", interval)
}