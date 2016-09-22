package src

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/toolkits/logger"

	"github.com/coraldane/DelayQ/g"
)

type DelayQueue struct {
}

func (this *DelayQueue) Publish(job Job) error {
	job.GmtCreate = time.Now()
	jobId := fmt.Sprintf("%d_%d_%d", job.Origin, job.Topic, job.BizId)

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	_, err := redis.Bool(rc.Do("HSET", g.JOB_INFO_KEY, jobId, job.String()))
	if nil != err {
		logger.Errorln("hset job error", job, err)
		return err
	}

	bucketKey := fmt.Sprintf("%s_%d", g.DELAY_BUCKET_PREFIX, CalcBucketId(jobId))
	_, err = redis.Int(rc.Do("ZADD", bucketKey, job.GmtNext.Unix(), jobId))
	if nil != err {
		logger.Errorln("put into delay bucket error", jobId)
		return err
	}
	return nil
}

func (this *DelayQueue) Consume(origin, topic int, nextSeconds int64) *Job {
	rc := g.RedisConnPool.Get()
	defer rc.Close()

	queueKey := fmt.Sprintf("%s_%d_%d", g.READY_QUEUE_PREFIX, origin, topic)
	jobId, err := redis.String(rc.Do("RPOP", queueKey))
	if nil != err || "" == jobId {
		return nil
	}

	bucketKey := fmt.Sprintf("%s_%d", g.WAIT_ACK_PREFIX, CalcBucketId(jobId))
	nextTime := time.Unix(time.Now().Unix()+nextSeconds, 0)
	reply, err := redis.Bool(rc.Do("ZADD", bucketKey, nextTime.Unix(), jobId))
	if nil != err {
		logger.Errorln("put job into wait ack bucket error", jobId, err)
		return nil
	} else if false == reply {
		logger.Errorln("job already exists in wait ack bucket", jobId)
		return nil
	}

	job := GetJobInfo(jobId)
	if nil != job {
		job.GmtNext = nextTime
		rc.Do("HSET", g.JOB_INFO_KEY, jobId, job.String())
	}
	return job
}

func (this *DelayQueue) Ack(jobId string) error {
	rc := g.RedisConnPool.Get()
	defer rc.Close()

	bucketKey := fmt.Sprintf("%s_%d", g.WAIT_ACK_PREFIX, CalcBucketId(jobId))
	reply, err := redis.Bool(rc.Do("ZREM", bucketKey, jobId))
	if nil != err {
		logger.Errorln("remove from ACK bucket error", jobId, err)
		return err
	}
	if reply {
		rc.Do("HDEL", g.JOB_INFO_KEY, jobId)
	}
	return nil
}
