package snowflake

import (
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	epoch          = int64(1735689600000)
	machineIDBits  = uint8(10)
	sequenceBits   = uint8(12)
	maxMachineID   = int64(-1) ^ (int64(-1) << machineIDBits)
	maxSequence    = int64(-1) ^ (int64(-1) << sequenceBits)
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
)

var (
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
	machineID     int64
)

func init() {
	machineID = getMachineID()
}

func getMachineID() int64 {
	machineIDStr := os.Getenv("MACHINE_ID")
	if machineIDStr == "" {
		return 1
	}
	id, err := strconv.ParseInt(machineIDStr, 10, 64)
	if err != nil || id < 0 || id > maxMachineID {
		return 1
	}
	return id
}

func currentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func waitNextMillis(last int64) int64 {
	ts := currentTimestamp()
	for ts <= last {
		ts = currentTimestamp()
	}
	return ts
}

func Generate() int64 {
	mu.Lock()
	defer mu.Unlock()

	ts := max(currentTimestamp(), lastTimestamp)

	if ts == lastTimestamp {
		sequence = (sequence + 1) & maxSequence
		if sequence == 0 {
			ts = waitNextMillis(lastTimestamp)
		}
	} else {
		sequence = 0
	}
	lastTimestamp = ts

	return ((ts - epoch) << timestampShift) |
		(machineID << machineIDShift) |
		sequence
}

func Parse(id int64) (timestamp int64, machineIDParsed int64, sequenceParsed int64) {
	timestamp = (id >> timestampShift) + epoch
	machineIDParsed = (id >> machineIDShift) & maxMachineID
	sequenceParsed = id & maxSequence
	return
}
