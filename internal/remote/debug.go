package remote

import "fmt"

var remoteVerbose bool

func SetVerbose(enabled bool) {
	remoteVerbose = enabled
}

func debugf(format string, args ...any) {
	if !remoteVerbose {
		return
	}
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}