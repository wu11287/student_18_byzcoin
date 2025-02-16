// Package main is an app to interact with an OmniLedger service. It can set up
// a new skipchain, store transactions and retrieve values given a key.
package main

import (
	"errors"
	"os"

	"student_18_byzcoin/omniledger/darc"
	"student_18_byzcoin/omniledger/service"
	// "github.com/dedis/student_18_omniledger/omniledger/darc"
	// "github.com/dedis/student_18_omniledger/omniledger/service"

	"gopkg.in/dedis/cothority.v2"
	"gopkg.in/dedis/kyber.v2/util/key"
	"gopkg.in/dedis/onet.v2/app"
	"gopkg.in/dedis/onet.v2/log"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "OmniLedger app"
	cliApp.Usage = "Key/value storage for OmniLedger"
	cliApp.Version = "0.1"
	cliApp.Commands = []cli.Command{
		{
			Name:      "create",
			Usage:     "creates a new skipchain",
			Aliases:   []string{"c"},
			ArgsUsage: "group.toml",
			Action:    create,
		},
	}
	cliApp.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
	}
	cliApp.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		return nil
	}
	log.ErrFatal(cliApp.Run(os.Args))
}

// Creates a new skipchain
func create(c *cli.Context) error {
	log.Info("Create a new skipchain")

	if c.NArg() != 1 {
		return errors.New("please give: group.toml")
	}
	group := readGroup(c)

	kp := key.NewKeyPair(cothority.Suite)

	client := service.NewClient()
	signer := darc.NewSignerEd25519(kp.Public, kp.Private)
	msg, err := service.DefaultGenesisMsg(service.CurrentVersion, group.Roster, []string{"Spawn_dummy"}, signer.Identity())
	if err != nil {
		return err
	}
	// tmp := group.Roster.List[0].URL
	// fmt.Println(tmp)
	resp, err := client.CreateGenesisBlock(group.Roster, msg)
	if err != nil {
		return errors.New("during creation of skipchain: " + err.Error())
	}
	log.Infof("Created new skipchain on roster %s with ID: %x", group.Roster.List, resp.Skipblock.Hash)
	log.Infof("Private: %s", kp.Private)
	log.Infof("Public: %s", kp.Public)
	return nil
}

// readGroup decodes the group given in the file with the name in the
// first argument of the cli.Context.
func readGroup(c *cli.Context) *app.Group {
	name := c.Args().First()
	f, err := os.Open(name)
	log.ErrFatal(err, "Couldn't open group definition file")
	group, err := app.ReadGroupDescToml(f)
	log.ErrFatal(err, "Error while reading group definition file", err)
	if len(group.Roster.List) == 0 {
		log.ErrFatalf(err, "Empty entity or invalid group defintion in: %s",
			name)
	}
	return group
}
