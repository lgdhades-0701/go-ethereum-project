package ethchain

import (
	"bytes"
	"fmt"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/ethereum/eth-go/ethwire"
	"math/big"
	"sync"
	"time"
)

type BlockProcessor interface {
	ProcessBlock(block *Block)
}

type EthManager interface {
	StateManager() *StateManager
	BlockChain() *BlockChain
	TxPool() *TxPool
	Broadcast(msgType ethwire.MsgType, data []interface{})
	Reactor() *ethutil.ReactorEngine
}

// TODO rename to state manager
type StateManager struct {
	// Mutex for locking the block processor. Blocks can only be handled one at a time
	mutex sync.Mutex

	// Canonical block chain
	bc *BlockChain
	// States for addresses. You can watch any address
	// at any given time
	addrStateStore *AddrStateStore

	// Stack for processing contracts
	stack *Stack
	// non-persistent key/value memory storage
	mem map[string]*big.Int

	Pow PoW

	Ethereum EthManager

	SecondaryBlockProcessor BlockProcessor

	// The managed states
	// Processor state. Anything processed will be applied to this
	// state
	procState *State
	// Comparative state it used for comparing and validating end
	// results
	compState *State

	// Mining state, solely used for mining
	miningState *State
}

func NewStateManager(ethereum EthManager) *StateManager {
	sm := &StateManager{
		stack:          NewStack(),
		mem:            make(map[string]*big.Int),
		Pow:            &EasyPow{},
		Ethereum:       ethereum,
		addrStateStore: NewAddrStateStore(),
		bc:             ethereum.BlockChain(),
	}

	return sm
}

func (sm *StateManager) ProcState() *State {
	return sm.procState
}

func (sm *StateManager) MiningState() *State {
	return sm.miningState
}

// Watches any given address and puts it in the address state store
func (sm *StateManager) WatchAddr(addr []byte) *AccountState {
	//FIXME account := sm.procState.GetAccount(addr)
	account := sm.bc.CurrentBlock.state.GetAccount(addr)

	return sm.addrStateStore.Add(addr, account)
}

func (sm *StateManager) GetAddrState(addr []byte) *AccountState {
	account := sm.addrStateStore.Get(addr)
	if account == nil {
		a := sm.bc.CurrentBlock.state.GetAccount(addr)
		account = &AccountState{Nonce: a.Nonce, Account: a}
	}

	return account
}

func (sm *StateManager) BlockChain() *BlockChain {
	return sm.bc
}

func (sm *StateManager) MakeContract(tx *Transaction) {
	contract := MakeContract(tx, sm.procState)
	if contract != nil {
		sm.procState.states[string(tx.Hash()[12:])] = contract.state
	}
}
func (sm *StateManager) ApplyTransaction(block *Block, tx *Transaction) {
}

func (sm *StateManager) ApplyTransactions(block *Block, txs []*Transaction) {
	// Process each transaction/contract
	for _, tx := range txs {
		// If there's no recipient, it's a contract
		if tx.IsContract() {
			//FIXME sm.MakeContract(tx)
			block.MakeContract(tx)
		} else {
			//FIXME if contract := procState.GetContract(tx.Recipient); contract != nil {
			if contract := block.state.GetContract(tx.Recipient); contract != nil {
				sm.ProcessContract(contract, tx, block)
			} else {
				err := sm.Ethereum.TxPool().ProcessTransaction(tx, block)
				if err != nil {
					ethutil.Config.Log.Infoln("[smGR]", err)
				}
			}
		}
	}
}

// The prepare function, prepares the state manager for the next
// "ProcessBlock" action.
func (sm *StateManager) Prepare(processer *State, comparative *State) {
	sm.compState = comparative
	sm.procState = processer
}

func (sm *StateManager) PrepareMiningState() {
	sm.miningState = sm.BlockChain().CurrentBlock.State()
}

// Default prepare function
func (sm *StateManager) PrepareDefault(block *Block) {
	sm.Prepare(sm.BlockChain().CurrentBlock.State(), block.State())
}

// Block processing and validating with a given (temporarily) state
func (sm *StateManager) ProcessBlock(block *Block) error {
	// Processing a blocks may never happen simultaneously
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	// Defer the Undo on the Trie. If the block processing happened
	// we don't want to undo but since undo only happens on dirty
	// nodes this won't happen because Commit would have been called
	// before that.
	defer sm.bc.CurrentBlock.Undo()

	hash := block.Hash()

	if sm.bc.HasBlock(hash) {
		return nil
	}

	// Check if we have the parent hash, if it isn't known we discard it
	// Reasons might be catching up or simply an invalid block
	if !sm.bc.HasBlock(block.PrevHash) && sm.bc.CurrentBlock != nil {
		return ParentError(block.PrevHash)
	}

	// Process the transactions on to current block
	sm.ApplyTransactions(sm.bc.CurrentBlock, block.Transactions())

	// Block validation
	if err := sm.ValidateBlock(block); err != nil {
		return err
	}

	// I'm not sure, but I don't know if there should be thrown
	// any errors at this time.
	if err := sm.AccumelateRewards(sm.bc.CurrentBlock, block); err != nil {
		return err
	}

	// if !sm.compState.Cmp(sm.procState)
	if !block.state.Cmp(sm.bc.CurrentBlock.state) {
		return fmt.Errorf("Invalid merkle root. Expected %x, got %x", block.State().trie.Root, sm.bc.CurrentBlock.State().trie.Root)
		//FIXME return fmt.Errorf("Invalid merkle root. Expected %x, got %x", sm.compState.trie.Root, sm.procState.trie.Root)
	}

	// Calculate the new total difficulty and sync back to the db
	if sm.CalculateTD(block) {
		// Sync the current block's state to the database and cancelling out the deferred Undo
		sm.bc.CurrentBlock.Sync()
		//FIXME sm.procState.Sync()

		// Broadcast the valid block back to the wire
		//sm.Ethereum.Broadcast(ethwire.MsgBlockTy, []interface{}{block.Value().Val})

		// Add the block to the chain
		sm.bc.Add(block)

		// If there's a block processor present, pass in the block for further
		// processing
		if sm.SecondaryBlockProcessor != nil {
			sm.SecondaryBlockProcessor.ProcessBlock(block)
		}

		ethutil.Config.Log.Infof("[smGR] Added block #%d (%x)\n", block.BlockInfo().Number, block.Hash())
		sm.Ethereum.Reactor().Post("newBlock", block)
	} else {
		fmt.Println("total diff failed")
	}

	return nil
}

func (sm *StateManager) CalculateTD(block *Block) bool {
	uncleDiff := new(big.Int)
	for _, uncle := range block.Uncles {
		uncleDiff = uncleDiff.Add(uncleDiff, uncle.Difficulty)
	}

	// TD(genesis_block) = 0 and TD(B) = TD(B.parent) + sum(u.difficulty for u in B.uncles) + B.difficulty
	td := new(big.Int)
	td = td.Add(sm.bc.TD, uncleDiff)
	td = td.Add(td, block.Difficulty)

	// The new TD will only be accepted if the new difficulty is
	// is greater than the previous.
	if td.Cmp(sm.bc.TD) > 0 {
		// Set the new total difficulty back to the block chain
		sm.bc.SetTotalDifficulty(td)

		return true
	}

	return false
}

// Validates the current block. Returns an error if the block was invalid,
// an uncle or anything that isn't on the current block chain.
// Validation validates easy over difficult (dagger takes longer time = difficult)
func (sm *StateManager) ValidateBlock(block *Block) error {
	// TODO
	// 2. Check if the difficulty is correct

	// Check each uncle's previous hash. In order for it to be valid
	// is if it has the same block hash as the current
	previousBlock := sm.bc.GetBlock(block.PrevHash)
	for _, uncle := range block.Uncles {
		if bytes.Compare(uncle.PrevHash, previousBlock.PrevHash) != 0 {
			return ValidationError("Mismatch uncle's previous hash. Expected %x, got %x", previousBlock.PrevHash, uncle.PrevHash)
		}
	}

	diff := block.Time - sm.bc.CurrentBlock.Time
	if diff < 0 {
		return ValidationError("Block timestamp less then prev block %v", diff)
	}

	// New blocks must be within the 15 minute range of the last block.
	if diff > int64(15*time.Minute) {
		return ValidationError("Block is too far in the future of last block (> 15 minutes)")
	}

	// Verify the nonce of the block. Return an error if it's not valid
	if !sm.Pow.Verify(block.HashNoNonce(), block.Difficulty, block.Nonce) {
		return ValidationError("Block's nonce is invalid (= %v)", block.Nonce)
	}

	return nil
}

func CalculateBlockReward(block *Block, uncleLength int) *big.Int {
	base := new(big.Int)
	for i := 0; i < uncleLength; i++ {
		base.Add(base, UncleInclusionReward)
	}
	return base.Add(base, BlockReward)
}

func CalculateUncleReward(block *Block) *big.Int {
	return UncleReward
}

func (sm *StateManager) AccumelateRewards(processor *Block, block *Block) error {
	// Get the coinbase rlp data
	addr := processor.state.GetAccount(block.Coinbase)
	//FIXME addr := proc.GetAccount(block.Coinbase)
	// Reward amount of ether to the coinbase address
	addr.AddFee(CalculateBlockReward(block, len(block.Uncles)))

	processor.state.UpdateAccount(block.Coinbase, addr)
	//FIXME proc.UpdateAccount(block.Coinbase, addr)

	for _, uncle := range block.Uncles {
		uncleAddr := processor.state.GetAccount(uncle.Coinbase)
		uncleAddr.AddFee(CalculateUncleReward(uncle))

		processor.state.UpdateAccount(uncle.Coinbase, uncleAddr)
		//FIXME proc.UpdateAccount(uncle.Coinbase, uncleAddr)
	}

	return nil
}

func (sm *StateManager) Stop() {
	sm.bc.Stop()
}

func (sm *StateManager) ProcessContract(contract *Contract, tx *Transaction, block *Block) {
	// Recovering function in case the VM had any errors
	/*
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from VM execution with err =", r)
			}
		}()
	*/

	vm := &Vm{}
	//vm.Process(contract, sm.procState, RuntimeVars{
	vm.Process(contract, block.state, RuntimeVars{
		address:     tx.Hash()[12:],
		blockNumber: block.BlockInfo().Number,
		sender:      tx.Sender(),
		prevHash:    block.PrevHash,
		coinbase:    block.Coinbase,
		time:        block.Time,
		diff:        block.Difficulty,
		txValue:     tx.Value,
		txData:      tx.Data,
	})
}
