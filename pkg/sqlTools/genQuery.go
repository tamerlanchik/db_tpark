package sqlTools

import (
	"bytes"
	"strconv"
	"strings"
)

func CreatePacketQuery(prefix string, batchSize int, batchCount int, postfix ...string) string {
	pack := make([]string, 0, batchCount)
	batch := make([]string, 0, batchSize)

	for i := 0; i < batchCount; i++ {
		for j := 1; j <= batchSize; j++ {
			batch = append(batch, "$"+strconv.Itoa(batchSize*i + j))
		}
		pack = append(pack, "(" + strings.Join(batch, ", ") + ")" )
		batch = batch[:0]
	}

	var res bytes.Buffer
	res.WriteString(prefix)
	if prefix[len(prefix)-1] != ' ' {
		res.WriteString(" ")
	}
	res.WriteString(strings.Join(pack, ", "))
	if len(postfix) > 0 {
		res.WriteString(" ")
		res.WriteString(strings.Join(postfix, " "))
	}
	res.WriteString(";")
	return res.String()
}
