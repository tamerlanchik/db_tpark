package structs

import "fmt"

func NoEmptyWrapper(def string, numb int) string{
	return fmt.Sprintf(`COALESCE(NULLIF($%d, ''), %s)`, numb, def)
}
