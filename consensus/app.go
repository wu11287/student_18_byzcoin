package consensus

import (
	"errors"
	"os"

	"gopkg.in/dedis/onet.v2/log"
	"gopkg.in/urfave/cli.v1"
)


func create(c *cli.Context) error {
	log.Info("create chain!")

	if c.NArg() != 1 {
		return errors.New("please give configure file : shard.txt ！")
	}



}

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


func readShardTxt(c *cli.Context)  {
	name := c.Args().First()
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}

}