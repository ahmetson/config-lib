package block

import (
	"strings"

	spaghetti_log "github.com/blocklords/sds/blockchain/event"
	"github.com/blocklords/sds/blockchain/evm/event"

	eth_types "github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	NetworkId      string
	BlockNumber    uint64
	BlockTimestamp uint64
	Logs           []*spaghetti_log.Log
}

func SetLogs(block *Block, raw_logs []eth_types.Log) error {
	var logs []*spaghetti_log.Log
	for _, rawLog := range raw_logs {
		if rawLog.Removed {
			continue
		}

		log := event.NewSpaghettiLog(block.NetworkId, block.BlockTimestamp, &rawLog)
		logs = append(logs, log)
	}

	block.Logs = logs

	return nil
}

// Returns the smartcontract information
// Todo Get the logs for the blockchain
// Rather than getting transactions
func (block *Block) GetForSmartcontract(address string) []*spaghetti_log.Log {
	logs := make([]*spaghetti_log.Log, 0)

	for _, log := range block.Logs {
		if strings.EqualFold(address, log.Address) {
			logs = append(logs, log)
		}
	}

	return logs
}
