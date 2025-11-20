// Abstraksi memory pool simple
package mempool

import (
	"sync"

	"github.com/bpjs-hackathon/sehat-chain/types"
)

type MemPool struct {
	mux          sync.RWMutex
	transactions map[string]types.Transaction // order berdasarkan hash
	order        []string                     // fifo ordering dari hash
}

func NewPool() *MemPool {
	return &MemPool{
		transactions: make(map[string]types.Transaction),
		order:        []string{},
	}
}

// Add tx ke mempool
func (mp *MemPool) AddTransaction(tx types.Transaction) {
	mp.mux.Lock()
	defer mp.mux.Unlock()

	// Transaction sudah ada
	if _, exists := mp.transactions[tx.ID]; exists {
		return
	}

	mp.transactions[tx.ID] = tx
	mp.order = append(mp.order, tx.ID)
}

// Ambil n total dari tx yang tersimpan di mempool
func (mp *MemPool) PickTxs(count int) []types.Transaction {
	mp.mux.RLock()
	defer mp.mux.RUnlock()

	if count > len(mp.order) {
		count = len(mp.order)
	}

	selected := make([]types.Transaction, 0, count)

	// Helper to pick from a specific queue
	picker := func(queue []string) {
		for _, txID := range queue {
			if len(selected) >= count {
				return
			}
			// Check if tx still exists (might have been removed by a committed block)
			if tx, exists := mp.transactions[txID]; exists {
				selected = append(selected, tx)
			}
		}
	}

	picker(mp.order)

	return selected
}

// RemoveTxs setelah block dibuat dan txs aman dibuang dari mempool
func (mp *MemPool) RemoveTxs(txs []types.Transaction) {
	mp.mux.Lock()
	defer mp.mux.Unlock()

	for _, tx := range txs {
		delete(mp.transactions, tx.ID) // <--- 4. CORRECT DELETION
	}

	mp.order = mp.rebuildSlice(mp.order)
}

// Filter key dari tx yang sudah dihapus
func (mp *MemPool) rebuildSlice(queue []string) []string {
	active := make([]string, 0, len(queue))

	for _, txID := range queue {
		if _, exists := mp.transactions[txID]; exists {
			active = append(active, txID)
		}
	}
	return active
}

// Hitung tx yang tersimpan di mempool
func (mp *MemPool) Count() int {
	mp.mux.RLock()
	defer mp.mux.RUnlock()

	return len(mp.transactions)
}
