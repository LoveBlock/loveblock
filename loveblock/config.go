// Copyright 2016 The loveblock Authors
// This file is part of the loveblock library.
//
// The loveblock library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The loveblock library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the loveblock library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/LoveBlock/loveblock/common/dpovp"
	"github.com/LoveBlock/loveblock/loveblock/utils"
	"github.com/LoveBlock/loveblock/network"
	"github.com/LoveBlock/loveblock/node"
	"github.com/LoveBlock/loveblock/params"
	"gopkg.in/urfave/cli.v1"
)

type loveblockConfig struct {
	Love network.Config
	Node node.Config
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "network")
	cfg.WSModules = append(cfg.WSModules, "network")
	cfg.IPCPath = "loveblock.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, loveblockConfig) {
	// Load defaults.
	cfg := loveblockConfig{
		Love: network.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetLoveConfig(ctx, stack, &cfg.Love)

	// sman 设置 datadir
	dpovp.SetDataDir(cfg.Node.DataDir)

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)

	utils.RegisterLoveService(stack, &cfg.Love)

	return stack
}
