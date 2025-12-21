package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	epoch          = int64(1735689600000) // Jan 1, 2025
	machineIDBits  = uint8(10)
	sequenceBits   = uint8(12)
	maxMachineID   = int64(-1) ^ (int64(-1) << machineIDBits)
	maxSequence    = int64(-1) ^ (int64(-1) << sequenceBits)
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
)

var (
	lastTimestamp int64
	sequence      uint32
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
	last := atomic.LoadInt64(&lastTimestamp)
	ts := currentTimestamp()

	if ts < last {
		panic(fmt.Sprintf("clock moved backwards: %d ms", last-ts))
	}

	if ts == last {
		seq := atomic.AddUint32(&sequence, 1) & uint32(maxSequence)
		if seq == 0 {
			ts = waitNextMillis(last)
			atomic.StoreUint32(&sequence, 0)
			atomic.StoreInt64(&lastTimestamp, ts)
		}
	} else {
		atomic.StoreUint32(&sequence, 0)
		atomic.StoreInt64(&lastTimestamp, ts)
	}

	seq := atomic.LoadUint32(&sequence)
	id := ((ts - epoch) << timestampShift) |
		(machineID << machineIDShift) |
		int64(seq)

	return id
}

func Parse(id int64) (timestamp int64, machineIDParsed int64, sequenceParsed int64) {
	timestamp = (id >> timestampShift) + epoch
	machineIDParsed = (id >> machineIDShift) & maxMachineID
	sequenceParsed = id & maxSequence
	return
}
