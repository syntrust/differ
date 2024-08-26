package main

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v3"
)

const (
	HEADER_SIZE       = 4096
	shardEntryBits    = 3 // from contract
	kvEntriesPerShard = 1 << shardEntryBits
	sampleSizeBits    = 5
	sampleSize        = 1 << sampleSizeBits
	kvSizeBits        = 17
	kvSize            = 1 << kvSizeBits
)

var maxUint256 = new(big.Int).Sub(
	new(big.Int).Exp(
		new(big.Int).SetUint64(2),
		new(big.Int).SetUint64(256),
		nil,
	),
	new(big.Int).SetUint64(1),
)

type mined struct {
	Input struct {
		Miner          string   `yaml:"miner"`
		Shard          uint64   `yaml:"shard"`
		Block          uint64   `yaml:"block"`
		BlockTime      uint64   `yaml:"blockTime"`
		LastMinedTime  uint64   `yaml:"lastMinedTime"`
		Difficulty     uint64   `yaml:"difficulty"`
		MixHash        string   `yaml:"mixHash"`
		Nonce          uint64   `yaml:"nonce"`
		Cutoff         uint64   `yaml:"cutoff"`
		DiffAdjDivisor uint64   `yaml:"diffAdjDivisor"`
		MinimumDiff    uint64   `yaml:"minimumDiff"`
		SampleIdxsInKv []uint64 `yaml:"sampleIdxsInKv"`
		KvIdxs         []uint64 `yaml:"kvIdxs"`
	} `yaml:"input"`
}

func (m mined) String() string {
	return fmt.Sprintf("Miner: %s\nShard: %d\nBlock: %d\nBlockTime: %d\nLastMineTime: %d\nDifficulty: %d\nMixHash: %s\nNonce: %d\nCutoff: %d\nDiffAdjDivisor: %d\nMinimumDiff: %d\nSampleIdxsInKv: %v\nKvIdxs: %v\n",
		m.Input.Miner, m.Input.Shard, m.Input.Block, m.Input.BlockTime, m.Input.LastMinedTime, m.Input.Difficulty, m.Input.MixHash, m.Input.Nonce, m.Input.Cutoff, m.Input.DiffAdjDivisor, m.Input.MinimumDiff, m.Input.SampleIdxsInKv, m.Input.KvIdxs)
}

func expectedDiff(lastMineTime, minedTime uint64, difficulty, cutoff, diffAdjDivisor, minDiff *big.Int) *big.Int {
	interval := new(big.Int).SetUint64(minedTime - lastMineTime)
	diff := difficulty
	if interval.Cmp(cutoff) < 0 {
		// diff = diff + (diff-interval*diff/cutoff)/diffAdjDivisor
		diff = new(big.Int).Add(
			diff,
			new(big.Int).Div(
				new(big.Int).Sub(
					diff,
					new(big.Int).Div(
						new(big.Int).Mul(
							interval,
							diff,
						),
						cutoff,
					)),
				diffAdjDivisor,
			),
		)
	} else {
		// dec := (interval*diff/cutoff - diff) / diffAdjDivisor
		dec := new(big.Int).Div(new(big.Int).Sub(new(big.Int).Div(new(big.Int).Mul(interval, diff), cutoff), diff), diffAdjDivisor)
		if new(big.Int).Add(dec, minDiff).Cmp(diff) > 0 {
			diff = minDiff
		} else {
			diff = new(big.Int).Sub(diff, dec)
		}
	}

	fmt.Printf("interval=%d, diffIn=%s, diffOut=%s\n", minedTime-lastMineTime, difficulty, diff)
	return diff
}

func hashimoto(hash0 common.Hash, shard uint64, sampleIdxs []uint64) common.Hash {

	filename := fmt.Sprintf("/Users/dl/code/es-node/es-data/shard-%d.dat", shard)
	file, err := os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf("Error reading the storage file: %v", err)
	}
	hash := hash0
	for i := 0; i < 2; i++ {
		md := make([]byte, sampleSize)
		n, err := file.ReadAt(md,
			HEADER_SIZE+int64(sampleIdxs[i]<<sampleSizeBits)-int64(kvEntriesPerShard*shard*kvSize))
		if err != nil {
			log.Fatalf("Error reading the storage file: %v", err)
		}
		if n != sampleSize {
			log.Fatalf("Error reading the storage file: %v", err)
		}
		fmt.Printf("sample = 0x%x\n", md)
		hash = crypto.Keccak256Hash(hash.Bytes(), md)
	}
	return hash
}

func main() {
	path := "/Users/dl/code/es-node/cmd/analyzer/mined1.yaml"
	configFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading the config file: %v", err)
	}

	var config mined
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling the config file: %v", err)
	}
	input := config.Input
	fmt.Printf("%s\n", config)
	hash0 := crypto.Keccak256Hash(
		common.HexToHash(input.Miner).Bytes(),
		common.FromHex(input.MixHash),
		common.BigToHash(new(big.Int).SetUint64(input.Nonce)).Bytes(),
	)
	var sampleIdxs []uint64
	for i := 0; i < 2; i++ {
		sampleIndex := input.KvIdxs[i]<<(kvSizeBits-sampleSizeBits) + input.SampleIdxsInKv[i]
		fmt.Printf("sampleIndex=%d \n", sampleIndex)
		sampleIdxs = append(sampleIdxs, sampleIndex)
	}
	hash := hashimoto(hash0, input.Shard, sampleIdxs)
	fmt.Printf("hash0=%x\n", hash0)
	fmt.Printf("hash =%x\n", hash)
	reqDiff := new(big.Int).Div(maxUint256,
		expectedDiff(
			input.LastMinedTime,
			input.BlockTime,
			big.NewInt(int64(input.Difficulty)),
			big.NewInt(int64(input.Cutoff)),
			big.NewInt(int64(input.DiffAdjDivisor)),
			big.NewInt(int64(input.MinimumDiff)),
		))
	fmt.Printf("diff=%s\n", reqDiff)
	fmt.Printf("hash=%s\n", hash.Big())
	if hash.Big().Cmp(reqDiff) < 0 {
		fmt.Println("Valid")
	} else {
		fmt.Println("Invalid")
	}
}
