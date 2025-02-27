package accesslist

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

var (
	CODE     = common.Hash(sha256.Sum256([]byte("code")))
	CODEHASH = common.Hash(sha256.Sum256([]byte("codeHash")))
	BALANCE  = common.Hash(sha256.Sum256([]byte("balance")))
	NONCE    = common.Hash(sha256.Sum256([]byte("nonce")))
	ALIVE    = common.Hash(sha256.Sum256([]byte("alive")))
)

type Byte52 [52]byte

type ALTuple map[Byte52]struct{}

func Combine(addr common.Address, hash common.Hash) Byte52 {
	var key Byte52
	copy(key[:], addr[:])
	copy(key[20:], hash[:])
	return key
}

func (tuple ALTuple) Add(addr common.Address, hash common.Hash) {
	key := Combine(addr, hash)
	if _, ok := tuple[key]; ok {
		return
	}
	tuple[key] = struct{}{}
}

func (tuple ALTuple) Contains(key Byte52) bool {
	_, ok := tuple[key]
	return ok
}

type RW_AccessLists struct {
	ReadAL  ALTuple
	WriteAL ALTuple
}

func NewRWAccessLists() *RW_AccessLists {
	return &RW_AccessLists{
		ReadAL:  make(ALTuple),
		WriteAL: make(ALTuple),
	}
}

func (RWAL RW_AccessLists) AddReadAL(addr common.Address, hash common.Hash) {
	RWAL.ReadAL.Add(addr, hash)
}

func (RWAL RW_AccessLists) AddWriteAL(addr common.Address, hash common.Hash) {
	RWAL.WriteAL.Add(addr, hash)
}

func (RWAL RW_AccessLists) HasConflict(other RW_AccessLists) bool {
	for key := range RWAL.ReadAL {
		if other.WriteAL.Contains(key) {
			return true
		}
	}
	for key := range RWAL.WriteAL {
		if other.WriteAL.Contains(key) {
			return true
		}
		if other.ReadAL.Contains(key) {
			return true
		}
	}
	return false
}

func (RWAL RW_AccessLists) Merge(other RW_AccessLists) {
	for key := range other.ReadAL {
		RWAL.ReadAL.Add(common.BytesToAddress(key[:20]), common.BytesToHash(key[20:]))
	}
	for key := range other.WriteAL {
		RWAL.WriteAL.Add(common.BytesToAddress(key[:20]), common.BytesToHash(key[20:]))
	}
}

func (RWAL RW_AccessLists) ToMarshal() RW_AccessLists_Marshal {
	readSet := make(map[common.Address][]string)
	writeSet := make(map[common.Address][]string)
	for key := range RWAL.ReadAL {
		addr := common.BytesToAddress(key[:20])
		hash := common.BytesToHash(key[20:])
		readSet[addr] = append(readSet[addr], decodeHash(hash))
	}
	for key := range RWAL.WriteAL {
		addr := common.BytesToAddress(key[:20])
		hash := common.BytesToHash(key[20:])
		writeSet[addr] = append(writeSet[addr], decodeHash(hash))
	}
	return RW_AccessLists_Marshal{
		ReadSet:  readSet,
		WriteSet: writeSet,
	}
}

func (RWAL RW_AccessLists) Equal(other RW_AccessLists) bool {
	if len(RWAL.ReadAL) != len(other.ReadAL) {
		return false
	}
	if len(RWAL.WriteAL) != len(other.WriteAL) {
		return false
	}
	for key := range RWAL.ReadAL {
		if !other.ReadAL.Contains(key) {
			return false
		}
	}
	for key := range RWAL.WriteAL {
		if !other.WriteAL.Contains(key) {
			return false
		}
	}
	return true
}

func (RWAL RW_AccessLists) ToJSON() string {
	str, _ := json.Marshal(RWAL.ToMarshal())
	return common.Bytes2Hex(str)
}

func decodeHash(hash common.Hash) string {
	switch hash {
	case CODE:
		return "code"
	case BALANCE:
		return "balance"
	case ALIVE:
		return "alive"
	case CODEHASH:
		return "codeHash"
	case NONCE:
		return "nonce"
	default:
		return hash.Hex()
	}
}

func encodeHash(str string) common.Hash {
	switch str {
	case "code":
		return CODE
	case "balance":
		return BALANCE
	case "alive":
		return ALIVE
	case "codeHash":
		return CODEHASH
	case "nonce":
		return NONCE
	default:
		return common.HexToHash(str)
	}
}

type RW_AccessLists_Marshal struct {
	ReadSet  map[common.Address][]string `json:"readSet"`
	WriteSet map[common.Address][]string `json:"writeSet"`
}

func NewRWAccessListsMarshal() *RW_AccessLists_Marshal {
	return &RW_AccessLists_Marshal{
		ReadSet:  make(map[common.Address][]string),
		WriteSet: make(map[common.Address][]string),
	}
}

func (RWALM RW_AccessLists_Marshal) ToRWAL() *RW_AccessLists {
	res := NewRWAccessLists()
	for addr, hashList := range RWALM.ReadSet {
		for _, hash := range hashList {
			res.ReadAL.Add(addr, encodeHash(hash))
		}
	}
	for addr, hashList := range RWALM.WriteSet {
		for _, hash := range hashList {
			res.WriteAL.Add(addr, encodeHash(hash))
		}
	}
	return res
}
