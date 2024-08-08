package lock

import "fmt"

func leaseName(id string) string {
	return fmt.Sprintf("upgrade-operator-%s", id)
}
