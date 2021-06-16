package main
// package pow

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
	// "os/exec"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

const MAXINT64 int64 = math.MaxInt64
const n int = 100
const s int = 1

var m int = int(math.Pow(2, float64(s)))
var lock sync.Mutex

type node struct {
	id      int
	pk      crypto.VrfPubkey
	sk      crypto.VrfPrivkey
	hashRes string
	pkList  [n]crypto.VrfPubkey
}

type pkAndId struct {
	id int
	pk crypto.VrfPubkey
}

func newNode(i int) *node {
	// nNode.id = i
	pk_tmp, sk_tmp := crypto.VrfKeygen()

	return &node {
		id:		i,
		pk:		pk_tmp,
		sk:		sk_tmp,
	}
}

func broadcastPK(ch []chan *pkAndId, nNode *node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.Lock()

	tmp := &pkAndId{id, nNode.pk}
	go func() {
		for i := 0; i < n; i++ {
			ch[i] <- tmp
		}
	}()

	lock.Unlock()
}

func storePk(ch chan *pkAndId, nNode []*node, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < n; i++ {
		tmp := <-ch
		nNode[id].pkList[tmp.id] = tmp.pk
	}
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Doshard(nNode *node, id int) {
	tmp := nNode.hashRes[len(nNode.hashRes)-s:]
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
		if nNode.hashRes[:5] != "00000" {
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

func main() {
// func powMain() float64 {
	var wg sync.WaitGroup
	// nodes := [n]node{}
	nodes := make([]*node, n)
	chsPK := make([]chan *pkAndId, n)
	var ips = [n]string{}

	for i := 0; i < n; i++ {
		chsPK[i] = make(chan *pkAndId)
		nodes[i] = newNode(i)
	}

	for i := 0; i < n; i++ {
		ips[i] = "192.168.0." + strconv.Itoa(i)
	}

	rand.Seed(time.Now().Unix())
	randomness, err := GenerateRandomBytes(10)
	if err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		wg.Add(2)
		go broadcastPK(chsPK, nodes[i], i, &wg)
		go storePk(chsPK[i], nodes, i, &wg)
	}

	wg.Wait()

	start := time.Now()

	// // 得到cpu使用率
	// command := `../shells/collect_cpu.sh`
	// cmd := exec.Command("/bin/bash", command)
	// err = cmd.Run()
	// if err != nil {
	// 	panic(err)
	// }

	// 得到系统负载
	// command := `../shells/collect_load.sh`
	// cmd := exec.Command("/bin/bash", command)
	// err = cmd.Run()
	// if err != nil {
	// 	panic(err)
	// }
	// go getCpuInfo()

	go getCpuLoad()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go hashCompute(ips[i], randomness, nodes[i].pk[:], nodes[i], &wg)
	}

	wg.Wait()

	for _, v := range nodes {
		Doshard(v, v.id)
	}

	in := time.Since(start)

	fmt.Println("time consumed: ", in)

	// return in.Seconds()
}


func getCpuInfo() {
    // CPU使用率
    for i:=0; i < 5; i++ {
        percent, _ := cpu.Percent(time.Second, false)
        fmt.Printf("cpu percent:%v\n", percent)
    }
}

func getCpuLoad() {
    info, _ := load.Avg()
	for i:=0; i < 5; i++{
		fmt.Printf("load: %v\n", info)
		time.Sleep(100*time.Millisecond)
	}
}