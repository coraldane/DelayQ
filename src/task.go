package src

import (
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/toolkits/logger"

	"github.com/coraldane/DelayQ/g"
)

func Schedule() {
	ticker1 := time.NewTicker(time.Duration(1) * time.Second)
	ticker2 := time.NewTicker(time.Duration(1) * time.Second)
	for {
		select {
		case <-ticker1.C:
			scanDelayBucket()
		case <-ticker2.C:
			scanWaitAckBucket()
		}
	}
}

func scanDelayBucket() {
	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for i := 0; i < g.Config().BucketKeySize; i++ {
		bucketKey := fmt.Sprintf("%s_%d", g.DELAY_BUCKET_PREFIX, i)
		doScanDelayBucket(rc, bucketKey)
	}
}

func doScanDelayBucket(rc redis.Conn, bucketKey string) {
	result := lockBucket(rc, bucketKey)
	defer unlockBucket(rc, bucketKey)

	if !result {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Errorln(err)
		}
	}()

	reply, err := redis.Strings(rc.Do("ZRANGEBYSCORE", bucketKey, "-inf", time.Now().Unix()))
	if nil != err {
		logger.Errorln("scan delay bucket error", err)
	} else {
		for _, jobId := range reply {
			queueId := jobId[0:strings.LastIndex(jobId, "_")]
			reply, err := redis.Int(rc.Do("LPUSH", fmt.Sprintf("%s_%s", g.READY_QUEUE_PREFIX, queueId), jobId))
			if nil == err || 1 == reply {
				rc.Do("ZREM", bucketKey, jobId)
			}
		}
	}
}

func scanWaitAckBucket() {
	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for i := 0; i < g.Config().BucketKeySize; i++ {
		bucketKey := fmt.Sprintf("%s_%d", g.WAIT_ACK_PREFIX, i)
		doScanWaitAckBucket(rc, bucketKey)
	}
}

func doScanWaitAckBucket(rc redis.Conn, bucketKey string) {
	result := lockBucket(rc, bucketKey)
	defer unlockBucket(rc, bucketKey)

	if !result {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Errorln(err)
		}
	}()

	reply, err := redis.Strings(rc.Do("ZRANGEBYSCORE", bucketKey, "-inf", time.Now().Unix()))
	if nil != err {
		logger.Errorln("scan wait ack bucket error", err)
	} else {
		for _, jobId := range reply {
			job := GetJobInfo(jobId)
			if nil == job {
				rc.Do("ZREM", bucketKey, jobId)
				continue
			}
			delayBucketKey := fmt.Sprintf("%s_%d", g.DELAY_BUCKET_PREFIX, CalcBucketId(jobId))

			//add it into delay bucket again
			reply, err := redis.Bool(rc.Do("ZADD", delayBucketKey, job.GmtNext.Unix(), jobId))
			if nil != err || false == reply {
				continue
			}

			//remove from wait ack bucket
			rc.Do("ZREM", bucketKey, jobId)

			job.RetryCount++
			rc.Do("HSET", g.JOB_INFO_KEY, jobId, job.String())
		}
	}
}

func lockBucket(rc redis.Conn, key string) bool {
	reply, err := redis.Int(rc.Do("HSETNX", g.LOCK_KEY, key, time.Now().Unix()+60))
	if nil != err {
		logger.Errorln("lock key error, ", key, err)
		return false
	}
	return 1 == reply
}

func unlockBucket(rc redis.Conn, key string) {
	rc.Do("HDEL", g.LOCK_KEY, key)
}
