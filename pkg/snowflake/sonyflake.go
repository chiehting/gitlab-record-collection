package snowflake

import (
	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("snowflake not created")
	}
}

// NextID get the snowflake ID once
func NextID() (id uint64, err error) {
	id, err = sf.NextID()
	return
}
