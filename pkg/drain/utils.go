package drain

import (
	"fmt"
)

func nodeUnschedulablePatch(desired bool) []byte {
	return []byte(fmt.Sprintf("{\"spec\":{\"unschedulable\":%t}}", desired))
}
