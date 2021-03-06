package util

import (
	"sync"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

type NonceCache interface {
	Serialize(account ethcmn.Address, fn func() error) error
	Sync(account ethcmn.Address, syncFn func() (uint64, error))

	Set(account ethcmn.Address, nonce int64)
	Get(account ethcmn.Address) (nonce int64, loaded bool)
	Incr(account ethcmn.Address) int64
	Decr(account ethcmn.Address) int64
}

func NewNonceCache() NonceCache {
	return &nonceCache{
		nonces: new(sync.Map),
		locks:  new(sync.Map),
		guard:  new(sync.Map),
	}
}

type nonceCache struct {
	nonces *sync.Map // map[ethcmn.Address]int64
	locks  *sync.Map // map[ethcmn.Address]*sync.RWMutex
	guard  *sync.Map
}

// Serialize serializes access to the nonce cache for all goroutines, all nonce increments should be done
// in this context. If a transaction increments nonce, but has not been submitted,
// it will have exclusive right to decrease nonce back for other transactions.
func (n nonceCache) Serialize(account ethcmn.Address, fn func() error) error {
	mux, _ := n.guard.LoadOrStore(account, new(sync.Mutex))
	mux.(*sync.Mutex).Lock()
	defer mux.(*sync.Mutex).Unlock()

	return fn()
}

func (n nonceCache) Get(account ethcmn.Address) (int64, bool) {
	lock, _ := n.locks.LoadOrStore(account, new(sync.RWMutex))
	lock.(*sync.RWMutex).RLock()
	defer lock.(*sync.RWMutex).RUnlock()

	nonce, loaded := n.nonces.LoadOrStore(account, int64(0))

	return nonce.(int64), loaded
}

func (n nonceCache) Set(account ethcmn.Address, nonce int64) {
	lock, _ := n.locks.LoadOrStore(account, new(sync.RWMutex))
	lock.(*sync.RWMutex).Lock()
	defer lock.(*sync.RWMutex).Unlock()

	n.nonces.Store(account, nonce)
}

func (n nonceCache) Incr(account ethcmn.Address) int64 {
	lock, _ := n.locks.LoadOrStore(account, new(sync.RWMutex))
	lock.(*sync.RWMutex).Lock()
	defer lock.(*sync.RWMutex).Unlock()

	v, _ := n.nonces.LoadOrStore(account, int64(0))
	nonce := v.(int64)
	nonce++
	n.nonces.Store(account, nonce)
	return nonce
}

func (n nonceCache) Decr(account ethcmn.Address) int64 {
	lock, _ := n.locks.LoadOrStore(account, new(sync.RWMutex))
	lock.(*sync.RWMutex).Lock()
	defer lock.(*sync.RWMutex).Unlock()

	v, _ := n.nonces.LoadOrStore(account, int64(0))
	nonce := v.(int64)
	nonce--
	n.nonces.Store(account, nonce)
	return nonce
}

func (n nonceCache) Sync(account ethcmn.Address, syncFn func() (uint64, error)) {
	lock, _ := n.locks.LoadOrStore(account, new(sync.RWMutex))
	lock.(*sync.RWMutex).Lock()
	defer lock.(*sync.RWMutex).Unlock()

	nonce, err := syncFn()
	if err == nil {
		n.nonces.Store(account, int64(nonce))
	}
}
