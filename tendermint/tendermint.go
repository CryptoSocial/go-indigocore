// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tendermint

import (
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tendermint/abci/types"
	tmcommon "github.com/tendermint/go-common"
	cfg "github.com/tendermint/go-config"
	p2p "github.com/tendermint/go-p2p"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmtypes "github.com/tendermint/tendermint/types"
)

// RunNode runs a tendermint node with an in-proc ABCI app
// Copied and modified from
// https://github.com/tendermint/tendermint/blob/master/node/node.go
func RunNode(config cfg.Config, app types.Application) *node.Node {
	// Wait until the genesis doc becomes available
	genDocFile := config.GetString("genesis_file")
	if !tmcommon.FileExists(genDocFile) {
		log.Infof("Waiting for genesis file %v...", genDocFile)
		for {
			time.Sleep(time.Second)
			if !tmcommon.FileExists(genDocFile) {
				continue
			}
			jsonBlob, err := ioutil.ReadFile(genDocFile)
			if err != nil {
				log.Fatalf("Couldn't read GenesisDoc file: %v", err)
			}
			genDoc := tmtypes.GenesisDocFromJSON(jsonBlob)
			if genDoc.ChainID == "" {
				log.Fatalf("Genesis doc %v must include non-empty chain_id", genDocFile)
			}
			config.Set("chain_id", genDoc.ChainID)
		}
	}

	// Create & start node
	n := newNodeDefault(config, app)

	protocol, address := node.ProtocolAndAddress(config.GetString("node_laddr"))
	l := p2p.NewDefaultListener(protocol, address, config.GetBool("skip_upnp"))
	n.AddListener(l)
	err := n.Start()
	if err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Run the RPC server.
	if config.GetString("rpc_laddr") != "" {
		_, err := n.StartRPC()
		if err != nil {
			log.Fatal(err)
		}
	}
	return n
}

// RunNodeForever runs a tendermint node with an in-proc ABCI app and waits for an exit signal
func RunNodeForever(config cfg.Config, app types.Application) {
	n := RunNode(config, app)
	// Sleep forever and then...
	tmcommon.TrapSignal(func() {
		n.Stop()
	})
}

func newNodeDefault(config cfg.Config, app types.Application) *node.Node {
	// Get PrivValidator
	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	return node.NewNode(config, privValidator, proxy.NewLocalClientCreator(app))
}