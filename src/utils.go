package src

import (
	"encoding/json"
	"hash/crc32"

	"github.com/garyburd/redigo/redis"
	"github.com/toolkits/logger"

	"github.com/coraldane/DelayQ/g"
)

func CalcBucketId(jobId string) uint32 {
	return crc32.ChecksumIEEE([]byte(jobId)) % uint32(g.Config().BucketKeySize)
}

func GetJobInfo(jobId string) *Job {
	rc := g.RedisConnPool.Get()
	defer rc.Close()

	reply, err := redis.Bytes(rc.Do("HGET", g.JOB_INFO_KEY, jobId))
	if nil != err {
		logger.Errorln("get job info error", jobId, err)
		return nil
	}

	var job Job
	err = json.Unmarshal(reply, &job)
	if nil != err {
		logger.Errorln("unmarshal job info error", string(reply), err)
	}
	return &job
}
