// implements for Dpovp consensus
package dpovp

import (
	"errors"
	"math/big"
	"time"

	"bytes"
	"fmt"

	"github.com/LoveBlock/loveblock/common"
	commonDpovp "github.com/LoveBlock/loveblock/common/dpovp"
	"github.com/LoveBlock/loveblock/consensus"
	"github.com/LoveBlock/loveblock/core/state"
	"github.com/LoveBlock/loveblock/core/types"
	"github.com/LoveBlock/loveblock/crypto"
	"github.com/LoveBlock/loveblock/log"
	"github.com/LoveBlock/loveblock/lovedb"
	"github.com/LoveBlock/loveblock/params"
	"github.com/LoveBlock/loveblock/rpc"
)

var (
	BlockReward *big.Int = big.NewInt(5e+18) // Block reward in wei for successfully mining a block
)

type Dpovp struct {
	config *params.DpovpConfig // Consensus engine configuration parameters
	db     lovedb.Database     // Database to store and retrieve snapshot checkpoints

	coinbase      common.Address // LoveBlock address of the signing key
	timeoutTime   int64          // 超时时间
	blockInternal int64          // 出块间隔
}

// 新增一个DPOVP共识机
func New(config *params.DpovpConfig, db lovedb.Database, coinbase common.Address) *Dpovp {
	// TODO
	conf := *config

	return &Dpovp{
		config:        &conf,
		db:            db,
		coinbase:      coinbase,
		timeoutTime:   config.Timeout,
		blockInternal: config.Sleeptime,
	}
}

// 设置coinbase
func (d *Dpovp) SetCoinbase(coinbase common.Address) {
	d.coinbase = coinbase
}

// Author implements consensus.Engine, returning the LoveBlock address recovered
// from the signature in the header's extra-data section.
// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (d *Dpovp) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (d *Dpovp) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	return d.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (d *Dpovp) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))
	log.Debug("start VerifyHeaders")
	go func() {
		for i, header := range headers {
			err := d.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (d *Dpovp) verifyHeader(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		log.Debug("verifyHeader: header.Number == nil")
		return consensus.ErrInvalidNumber
	}
	if header.Difficulty.Uint64() != uint64(1) {
		return fmt.Errorf(`verifyHeader: block difficulty doesn't equal 1.`)
	}
	number := header.Number.Uint64()
	hashTmp := header.Hash()
	log.Debug(fmt.Sprintf("start verifyHeader: hash:%s num:%d", common.ToHex(hashTmp[:]), number))
	var parent *types.Header
	for _, b := range parents {
		if b.Hash() == header.ParentHash {
			parent = b
			break
		}
	}
	if parent == nil {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		log.Debug("verifyHeader: parent == nil")
		return consensus.ErrUnknownAncestor
	}
	// Don't waste time checking blocks from the future
	if header.Time.Cmp(big.NewInt(time.Now().Unix())) > 0 {
		log.Debug("verifyHeader: header.Time > time.Now()")
		return consensus.ErrFutureBlock
	}
	// 验证签名
	tmpHash := crypto.Keccak256Hash(header.Coinbase[:])
	pubKey, err := crypto.Ecrecover(tmpHash[:], header.SignInfo)
	if err != nil {
		return fmt.Errorf(`verifyHeader: Wrong signinfo`)
	}
	blkNodePubkey := commonDpovp.GetPubkeyByAddress(&(header.Coinbase)) // 获取出块者的node公钥
	if blkNodePubkey == nil {
		return fmt.Errorf("verifyHeader: Verify header failed. Cann't get pubkey of %s", common.ToHex(header.Coinbase[:]))
	}
	if bytes.Compare(blkNodePubkey, pubKey[1:]) != 0 {
		return fmt.Errorf("verifyHeader: Cann't verify block's signer")
	}

	// 以下为确定是否该该节点出块
	if parent.Number.Uint64() == uint64(0) { // 父块为创世块
		log.Debug("verifyHeader: parent is genesis block")
		return nil
	}
	timeSpan := int64(header.Time.Uint64()-parent.Time.Uint64()) * 1000 // 当前块与父块时间间隔 单位：ms
	if timeSpan < d.blockInternal {                                     // 块与父块的时间间隔至少为 block internal
		log.Debug(fmt.Sprintf("verifyHeader: timeSpan:%d is smaller than blockInternal:%d", timeSpan, d.blockInternal))
		return fmt.Errorf("verifyHeader: block is not enough newer than it's parent")
	}
	nodeCount := commonDpovp.GetCoreNodesCount() // 总节点数
	slot := commonDpovp.GetSlot(&(parent.Coinbase), &(header.Coinbase))
	oneLoopTime := int64(commonDpovp.GetCoreNodesCount()) * d.timeoutTime // 一轮全部超时时的时间
	log.Debug(fmt.Sprintf("verifyHeader: timeSpan:%d nodeCount:%d slot:%d oneLoopTime:%d", timeSpan, nodeCount, slot, oneLoopTime))
	// 只有一个出块节点
	if nodeCount == 1 {
		if timeSpan < d.blockInternal { // 块间隔至少blockInternal
			log.Debug("verifyHeader: Only one node, but not sleep enough time -1")
			return fmt.Errorf("verifyHeader: Only one node, but not sleep enough time -1")
		}
		log.Debug("verifyHeader: nodeCount == 1")
		return nil
	}

	if slot == 0 { // 上一个块为自己出的块
		timeSpan = timeSpan % oneLoopTime
		log.Debug(fmt.Sprintf("verifyHeader: slot:0 timeSpan:%d", timeSpan))
		if timeSpan >= oneLoopTime-d.timeoutTime {
			// 正常情况
		} else {
			log.Debug(fmt.Sprintf("verifyHeader: slot:0 verify failed"))
			return fmt.Errorf("verifyHeader: Not turn to produce block -2")
		}
		return nil
	} else if slot == 1 {
		if timeSpan < oneLoopTime { // 间隔不到一个循环
			if timeSpan >= d.blockInternal && timeSpan < d.timeoutTime {
				// 正常情况
			} else {
				log.Debug(fmt.Sprintf("verifyHeader: slot:1, timeSpan<oneLoopTime, verify failed"))
				return fmt.Errorf("verifyHeader: Not turn to produce block -3")
			}
		} else { // 间隔超过一个循环
			timeSpan = timeSpan % oneLoopTime
			if timeSpan < d.timeoutTime {
				// 正常情况
			} else {
				log.Debug(fmt.Sprintf("verifyHeader: slot:1,timeSpan>=oneLoopTime, verify failed"))
				return fmt.Errorf("verifyHeader: Not turn to produce block -4")
			}
		}
	} else {
		timeSpan = timeSpan % oneLoopTime
		log.Debug(fmt.Sprintf("verifyHeader: slot:%d timeSpan:%d", slot, timeSpan))
		if timeSpan/d.timeoutTime == int64(slot-1) {
			// 正常情况
		} else {
			log.Debug(fmt.Sprintf("verifyHeader: slot>1, verify failed"))
			return fmt.Errorf("verifyHeader: Not turn to produce block -5")
		}
	}
	return nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (d *Dpovp) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (d *Dpovp) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	log.Debug("VerifySeal: start VerifySeal")
	// 验证签名
	tmpHash := crypto.Keccak256Hash(header.Coinbase[:])
	pubkey, err := crypto.Ecrecover(tmpHash[:], header.SignInfo)
	if err != nil {
		hashTmp := header.Hash()
		return fmt.Errorf("VerifySeal: Failed to verify Seal. hash:%s", common.ToHex(hashTmp[:]))
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
	if bytes.Compare(header.Coinbase[:], signer[:]) != 0 {
		return fmt.Errorf(`VerifySeal: signer != coinbase`)
	}
	log.Debug("VerifySeal: end VerifySeal")
	return nil
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (d *Dpovp) Prepare(chain consensus.ChainReader, header *types.Header) error {
	log.Debug("mine-Prepare: start Prepare")
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Nonce is reserved for now, set to empty
	header.Nonce = types.BlockNonce{}
	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}
	// Set the difficulty to 1
	header.Difficulty = new(big.Int).SetInt64(1)
	header.Time = new(big.Int).SetUint64(uint64(time.Now().Unix()))
	log.Debug("mine-Prepare: end Prepare")
	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (d *Dpovp) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	log.Debug("mine-Finalize: start")
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	accumulateRewards(state, header)
	header.Root = state.IntermediateRoot(true)
	header.UncleHash = types.CalcUncleHash(nil)

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts), nil
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (d *Dpovp) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	// 判断本节点是否在主节点列表中
	coinbaseIndex := commonDpovp.GetCoreNodeIndex(&(d.coinbase))
	if coinbaseIndex == -1 {
		log.Debug(fmt.Sprintf("mine-Seal: coinbaseIndex==-1 coinbase:%s", common.ToHex(d.coinbase[:])))
		return nil, fmt.Errorf("Coinbase is not in star list.")
	}
	log.Debug("mine-Seal: start")
	// 出块
	header := block.Header()
	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return nil, fmt.Errorf("mine-Seal: unknownblock, number:%d", number)
	}
	// 对区块进行签名
	hash := crypto.Keccak256Hash(d.coinbase[:])
	privKey := commonDpovp.GetPrivKey()
	if signInfo, err := crypto.Sign(hash[:], &privKey); err != nil {
		log.Warn("mine-Seal: sign failed")
		return nil, err
	} else {
		header.SignInfo = make([]byte, len(signInfo))
		copy(header.SignInfo, signInfo)
	}
	result := block.WithSeal(header)
	log.Debug("mine-Seal: end")
	return result, nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (d *Dpovp) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int).SetInt64(1)
}

// APIs returns the RPC APIs this consensus engine provides.
func (d *Dpovp) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward
func accumulateRewards(state *state.StateDB, header *types.Header) {
	blockReward := BlockReward
	reward := new(big.Int).Set(blockReward)
	state.AddBalance(header.Coinbase, reward)
}

// NewTester creates a small sized DPoVP scheme useful only for testing purposes.
func NewTester() *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}

// NewFaker creates a dpovp consensus engine with a fake DPoVP scheme that accepts
// all blocks' seal as valid, though they still have to conform to the LoveBlock
// consensus rules.
func NewFaker() *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}

// NewFakeFailer creates a dpovp consensus engine with a fake DPoVP scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the LoveBlock consensus rules.
func NewFakeFailer(fail uint64) *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}

// NewFakeDelayer creates a dpovp consensus engine with a fake DPoVP scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the LoveBlock consensus rules.
func NewFakeDelayer(delay time.Duration) *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}

// NewFullFaker creates an dpovp consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}

// NewShared creates a full sized DPoVP shared between all requesters running
// in the same process.
func NewShared() *Dpovp {
	return &Dpovp{
		config: &params.DpovpConfig{},
	}
}
