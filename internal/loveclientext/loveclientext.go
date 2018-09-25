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

// package loveclientext contains loveblock specific network-client.js extensions.
package loveclientext

var Modules = map[string]string{
	"admin":    Admin_JS,
	"debug":    Debug_JS,
	"network":  Love_JS,
	"miner":    Miner_JS,
	"net":      Net_JS,
	"personal": Personal_JS,
	"rpc":      RPC_JS,
	"txpool":   TxPool_JS,
}

const Admin_JS = `
networkClient._extend({
	property: 'admin',
	methods: [
		new networkClient._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new networkClient._extend.Method({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new networkClient._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new networkClient._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new networkClient._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	],
	properties: [
		new networkClient._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new networkClient._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new networkClient._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const Debug_JS = `
networkClient._extend({
	property: 'debug',
	methods: [
		new networkClient._extend.Method({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new networkClient._extend.Method({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new networkClient._extend.Method({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new networkClient._extend.Method({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new networkClient._extend.Method({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new networkClient._extend.Method({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new networkClient._extend.Method({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new networkClient._extend.Method({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new networkClient._extend.Method({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new networkClient._extend.Method({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new networkClient._extend.Method({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'setMutexProfileRate',
			call: 'debug_setMutexProfileRate',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new networkClient._extend.Method({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new networkClient._extend.Method({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [null, null]
		}),
		new networkClient._extend.Method({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new networkClient._extend.Method({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new networkClient._extend.Method({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new networkClient._extend.Method({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new networkClient._extend.Method({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new networkClient._extend.Method({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new networkClient._extend.Method({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
	],
	properties: []
});
`

const Love_JS = `
networkClient._extend({
	property: 'network',
	methods: [
		new networkClient._extend.Method({
			name: 'sign',
			call: 'network_sign',
			params: 2,
			inputFormatter: [networkClient._extend.formatters.inputAddressFormatter, null]
		}),
		new networkClient._extend.Method({
			name: 'resend',
			call: 'network_resend',
			params: 3,
			inputFormatter: [networkClient._extend.formatters.inputTransactionFormatter, networkClient._extend.utils.fromDecimal, networkClient._extend.utils.fromDecimal]
		}),
		new networkClient._extend.Method({
			name: 'signTransaction',
			call: 'network_signTransaction',
			params: 1,
			inputFormatter: [networkClient._extend.formatters.inputTransactionFormatter]
		}),
		new networkClient._extend.Method({
			name: 'submitTransaction',
			call: 'network_submitTransaction',
			params: 1,
			inputFormatter: [networkClient._extend.formatters.inputTransactionFormatter]
		}),
		new networkClient._extend.Method({
			name: 'getRawTransaction',
			call: 'network_getRawTransactionByHash',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (networkClient._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'network_getRawTransactionByBlockHashAndIndex' : 'network_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [networkClient._extend.formatters.inputBlockNumberFormatter, networkClient._extend.utils.toHex]
		}),
	],
	properties: [
		new networkClient._extend.Property({
			name: 'pendingTransactions',
			getter: 'network_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(networkClient._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	]
});
`

const Miner_JS = `
networkClient._extend({
	property: 'miner',
	methods: [
		new networkClient._extend.Method({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new networkClient._extend.Method({
			name: 'stop',
			call: 'miner_stop'
		}),
		new networkClient._extend.Method({
			name: 'setLovebase',
			call: 'miner_setLovebase',
			params: 1,
			inputFormatter: [networkClient._extend.formatters.inputAddressFormatter]
		}),
		new networkClient._extend.Method({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new networkClient._extend.Method({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [networkClient._extend.utils.fromDecimal]
		}),
		new networkClient._extend.Method({
			name: 'getHashrate',
			call: 'miner_getHashrate'
		}),
	],
	properties: []
});
`

const Net_JS = `
networkClient._extend({
	property: 'net',
	methods: [],
	properties: [
		new networkClient._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const Personal_JS = `
networkClient._extend({
	property: 'personal',
	methods: [
		new networkClient._extend.Method({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, networkClient._extend.formatters.inputAddressFormatter, null]
		}),
		new networkClient._extend.Method({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		new networkClient._extend.Method({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new networkClient._extend.Method({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [networkClient._extend.formatters.inputTransactionFormatter, null]
		}),
	],
	properties: [
		new networkClient._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RPC_JS = `
networkClient._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new networkClient._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const TxPool_JS = `
networkClient._extend({
	property: 'txpool',
	methods: [],
	properties:
	[
		new networkClient._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new networkClient._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new networkClient._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = networkClient._extend.utils.toDecimal(status.pending);
				status.queued = networkClient._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`
