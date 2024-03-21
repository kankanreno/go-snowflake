// Package snowflake is a network service for generating unique ID numbers at high scale with some simple guarantees.
// The first bit is unused sign bit.
// The second part consists of a 41-bit timestamp (milliseconds) whose value is the offset of the current time relative to a certain time.
// The 5 bits of the third and fourth parts represent data center and worker, and max value is 2^5 -1 = 31.
// The last part consists of 12 bits, its means the length of the serial number generated per millisecond per working node, a maximum of 2^12 -1 = 4095 IDs can be generated in the same millisecond.
// In a distributed environment, five-bit datacenter and worker mean that can deploy 31 datacenters, and each datacenter can deploy up to 31 nodes.
// The binary length of 41 bits is at most 2^41 -1 millisecond = 69 years. So the snowflake algorithm can be used for up to 69 years, In order to maximize the use of the algorithm, you should specify a start time for it.
package snowflake

import (
	"errors"
	"time"
)

// These constants are the bit lengths of snowflake ID parts.
const (
	TimestampLength = 41
	MachineIDLength = 6
	SequenceLength  = 6
	MaxTimestamp    = 1<<TimestampLength - 1
	MaxMachineID    = 1<<MachineIDLength - 1
	MaxSequence     = 1<<SequenceLength - 1

	machineIDMoveLength = SequenceLength
	timestampMoveLength = MachineIDLength + SequenceLength
)

// SequenceResolver the snowflake sequence resolver.
//
// When you want use the snowflake algorithm to generate unique ID, You must ensure: The sequence-number generated in the same millisecond of the same node is unique.
// Based on this, we create this interface provide following reslover:
//
//	AtomicResolver : base sync/atomic (by default).
type SequenceResolver func(ms int64) (uint16, error)

// default machineID is 0
// default resolver is AtomicResolver
// default startTime is 2020-01-01 00:00:00 UTC
var (
	resolver  SequenceResolver
	machineID = 0
	startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ID use ID to generate snowflake id and it will ignore error. if you want error info, you need use NextID method.
// This function is thread safe.
func ID() uint64 {
	id, _ := NextID()
	return id
}

// NextID use NextID to generate snowflake id and return an error.
// This function is thread safe.
func NextID() (uint64, error) {
	c := currentMillis()
	seqResolver := callSequenceResolver()
	seq, err := seqResolver(c)

	if err != nil {
		return 0, err
	}

	for seq >= MaxSequence {
		c = waitForNextMillis(c)
		seq, err = seqResolver(c)
		if err != nil {
			return 0, err
		}
	}

	df := int(elapsedTime(c, startTime))
	if df < 0 || df > MaxTimestamp {
		return 0, errors.New("The maximum life cycle of the snowflake algorithm is 2^41(millis), please check starttime")
	}

	id := uint64((df << timestampMoveLength) | (machineID << machineIDMoveLength) | int(seq))
	return id, nil
}

// SetStartTime set the start time for snowflake algorithm.
//
// It will panic when:
//
// s IsZero
// s > current millisecond
// current millisecond - s > 2^41(69 years).
//
// This function is thread-unsafe, recommended you call him in the main function.
func SetStartTime(s time.Time) {
	s = s.UTC()

	if s.IsZero() {
		panic("The start time cannot be a zero value")
	}

	if s.After(time.Now().UTC()) {
		panic("The s cannot be greater than the current millisecond")
	}

	// Because s must after now, so the `df` not < 0.
	df := elapsedTime(currentMillis(), s)
	if df > MaxTimestamp {
		panic("The maximum life cycle of the snowflake algorithm is 69 years")
	}

	startTime = s
}

// SetMachineID specify the machine ID. It will panic when machineid > max limit for 2^6-1=63.
// This function is thread-unsafe, recommended you call him in the main function.
func SetMachineID(m uint16) {
	if m > MaxMachineID {
		panic("The machineid cannot be greater than 63")
	}
	machineID = int(m)
}

// SetSequenceResolver set an custom sequence resolver.
// This function is thread-unsafe, recommended you call him in the main function.
func SetSequenceResolver(seq SequenceResolver) {
	if seq != nil {
		resolver = seq
	}
}

// SID snowflake id
type SID struct {
	Sequence  uint64
	MachineID uint64
	Timestamp uint64
	ID        uint64
}

// GenerateTime snowflake generate at, return a UTC time.
func (id *SID) GenerateTime() time.Time {
	ms := startTime.UTC().UnixNano()/1e6 + int64(id.Timestamp)

	return time.Unix(0, ms*int64(time.Millisecond)).UTC()
}

// ParseID parse snowflake it to SID struct.
func ParseID(id uint64) SID {
	time := id >> (SequenceLength + MachineIDLength)
	sequence := id & MaxSequence
	machineID := (id & (MaxMachineID << SequenceLength)) >> SequenceLength

	return SID{
		ID:        id,
		Sequence:  sequence,
		MachineID: machineID,
		Timestamp: time,
	}
}

//--------------------------------------------------------------------
// private function defined.
//--------------------------------------------------------------------

func waitForNextMillis(last int64) int64 {
	now := currentMillis()
	for now == last {
		now = currentMillis()
	}
	return now
}

func callSequenceResolver() SequenceResolver {
	if resolver == nil {
		return AtomicResolver
	}

	return resolver
}

func elapsedTime(nowms int64, s time.Time) int64 {
	return nowms - s.UTC().UnixNano()/1e6
}

// currentMillis get current millisecond.
func currentMillis() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
