package idgen

import (
	"fmt"
	"sync"
	"time"
)

const (
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	// You may customize this to set a different epoch for your application.
	Epoch int64 = 1288834974657
	// NodeBits holds the number of bits to use for Snowflake
	// Remember, you have a total 22 bits to share between Snowflake/Step
	NodeBits uint8 = 10

	// StepBits holds the number of bits to use for Step
	// Remember, you have a total 22 bits to share between Snowflake/Step
	StepBits uint8 = 12

	nodeMax   uint64 = 1<<NodeBits - 1
	nodeMask         = nodeMax << StepBits
	stepMask  uint64 = 1<<StepBits - 1
	timeShift        = NodeBits + StepBits
	nodeShift        = StepBits
)

// A JSONSyntaxError is returned from UnmarshalJSON if an invalid ID is provided.
type JSONSyntaxError struct{ original []byte }

func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// A Snowflake struct holds the basic information needed for a snowflake generator
// node
type Snowflake struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64
	node  uint16
	step  uint16

	nodeMax   uint16
	nodeMask  uint16
	stepMask  uint16
	timeShift uint8
	nodeShift uint8
}

// NewSnowflake returns a new snowflake node that can be used to generate snowflake
// IDs
func NewSnowflake(node uint16, nodeBits uint8) *Snowflake {
	stepBits := 22 - nodeBits
	n := Snowflake{
		epoch:     time.Unix(Epoch/1000, (Epoch%1000)*1000000),
		time:      0,
		node:      node,
		step:      0,
		nodeMax:   1<<nodeBits - 1,
		nodeMask:  (1<<nodeBits - 1) << stepBits,
		stepMask:  1<<stepBits - 1,
		timeShift: nodeBits + stepBits,
		nodeShift: stepBits,
	}

	return &n
}

func (n *Snowflake) Epoch(t time.Time) *Snowflake {
	n.epoch = t
	return n
}

// Generate creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (n *Snowflake) Generate() uint64 {

	n.mu.Lock()

	now := time.Since(n.epoch).Nanoseconds() / 1000000

	if now == n.time {
		n.step = (n.step + 1) & n.stepMask

		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := uint64(now)<<n.timeShift |
		(uint64(n.node) << n.nodeShift) |
		uint64(n.step)

	n.mu.Unlock()
	return r
}

func (n *Snowflake) Decompose(id uint64) (timestamp int64, nodeId uint16, step uint16) {
	timestamp = int64(id>>n.timeShift) + n.epoch.UnixNano()/1000000
	nodeId = uint16(id>>n.nodeShift) & n.nodeMask
	step = uint16(id)
	return timestamp, nodeId, step
}
