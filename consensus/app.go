package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	initial "myProject/initial"
	crypto "myProject/crypto"
	"github.com/BurntSushi/toml"

	"gopkg.in/dedis/onet.v2/log"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "POS_VRF app"
	cliApp.Usage = "data storage"
	cliApp.Commands = []cli.Command {
		{
			Name: 	"create",
			Usage:  "create a shard chain",
			Aliases: []string{"c"},
			ArgsUsage: "shard.txt",
			Action: create,
		},
	}
	log.ErrFatal(cliApp.Run(os.Args))
}


// create a new chain
func create(c *cli.Context) error {
	log.Info("create chain!")

	if c.NArg() != 1 {
		return errors.New("please give configure file : shard.txt ！")
	}
	shard, err := readShard(c)
	if err != nil {
		log.ErrFatal(err, "read shard error!")
	}

	return nil
}


type Shard struct {
	List *NodesInShard
}

type NodesInShard struct {
	Id 			int
	Pk 			crypto.VrfPubkey
	Sk 			crypto.VrfPrivkey
	URL 		string
	Description string
}

//如何根据节点id 得到 nodes信息？？
func readShard(c *cli.Context) (*Shard, error) {
	log.Info("readShard")
	name := c.Args().First()
	f, err := os.Open(name)
	log.ErrFatal(err, "error while open shard txt")

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		log.ErrFatal(err, "error while read shard txt")
		return nil, err
	}
	
	var entities = make([]*NodesInShard, len(bs))
	for i, s := range bs {
		entities[i].Id = i
		entities[i].Pk = initial.Node.Pk
	}
}
	