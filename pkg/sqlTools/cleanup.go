package sqlTools

import "time"

func FormatDate(bd int, date time.Time) string {
	switch bd {
	case BDPostgres:
		return date.Format("2006-01-02 15:04:05")
	default:
		return ""
	}
}
