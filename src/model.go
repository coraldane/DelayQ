package src

import (
	"encoding/json"
	"strings"
	"time"
)

type myTime struct {
	time.Time
}

func (t *myTime) UnmarshalJSON(buf []byte) error {
	tt, err := time.Parse("2006-01-02 15:04:05", strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

type Job struct {
	Origin    int
	Topic     int
	BizId     int64
	GmtCreate time.Time
	GmtNext   time.Time

	Body       string
	RetryCount int
}

func (this *Job) String() string {
	bs, _ := json.Marshal(this)
	return string(bs)
}
