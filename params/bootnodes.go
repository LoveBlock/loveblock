// Copyright 2015 The loveblock Authors
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

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main LoveBlock network.
var MainnetBootnodes = []string{
	// LoveBlock Foundation Go Bootnodes
	"enode://948cee7ebfbb76f425344e77314dabd7146e6f14d1b81b7d10485cb120f8d7cc04687a193a4500a014fbd26ed9c6cb8c52b1d08b13c1c27d0574cb73c22517c6@149.28.68.93:60603",
	//"enode://84697cd4b36a9ef2cb84e6aab0b96096403d3ef7752e3b36bd68e40ab4dcdab5e2913c31044be600383fd9273acfe3d74633b496901014fd2ce88d40a7651c08@149.28.68.93:60604",
	//"enode://6ff44a2dd0a0d133cd301c68a4008eef0be6326dabbfcec9c3b709d6451090c419de355b461c69a571f11a6265e92fc68a6d996647126882df2c872c313917d5@149.28.68.93:60605",
	//"enode://971d00a674273c96940832179fb2ecd936ba49c33e1b37ebbc709acf3aa98171b33d458db76fbee55e629a0c074f2f7b3518225817d6862047bfd43da2fac102@149.28.25.8:60606",
	//"enode://ed8cd664a9d6f63f7b447884db5fc57c42362f03a83ef12db2122bd418dcf24177935a9f1108ec49323e8602f7c32febec1491cec3f991b22c098c9cc6a21e1e@45.77.121.107:60607",
	//"enode://1168662f58415f37b08a72238a6d52dd721e9707232be0fe21c22884082f074e815a0f7cb9255b42b9eb2341e5a7016e4f483aa626e145692706d5efd7310bf5@63.211.111.245:60608",
	//// bootnode
	"enode://46880c2bc68d7fcf6819d04bcbc626caa170fad970a3a2158dc5b6c92244ea786c8c41e9fa206194a9c96b579f28e903d13e0f057cf642dfc3f49028a0bfe419@149.28.68.93:60000",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://7b1cb5d049aa08ac6e7be442d44fe0c9d3e2404b3a6d71d3ebe14d5aa961bf0d232571c8d64dd0ceeff37208312f1591ff9b2bab8030b2dbb03b588e41f36ba9@149.28.68.93:50301",
}
