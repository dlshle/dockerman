package randx

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

func Int32() int32 {
	return r.Int31()
}
