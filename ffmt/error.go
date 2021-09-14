package ffmt

import "log"

func Err(s string, args ...interface{}) {
	log.Printf(s+"\n", args...)
}
