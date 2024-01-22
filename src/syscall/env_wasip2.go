//go:build wasip2

package syscall

import (
	"github.com/ydnar/wasm-tools-go/wasi/cli/environment"
)

func Environ() []string {
	var env []string
	for _, kv := range environment.GetEnvironment().Slice() {
		env = append(env, kv.V0+"="+kv.V1)
	}
	return env
}
