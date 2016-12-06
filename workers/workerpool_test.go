package workers

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	lock := sync.Mutex{}
	jobsList := []int64{}
	jobsRan := map[int64]int{}
	p := NewWokerPool(1, func(j *Job) (interface{}, error) {
		fmt.Printf("Job: %#v\n", j)
		/////
		lock.Lock()
		if v, ok := jobsRan[j.ID]; ok {
			jobsRan[j.ID] = v + 1
		} else {
			jobsRan[j.ID] = 1
		}
		jobsList = append(jobsList, j.ID)
		lock.Unlock()
		/////
		return "here", nil
	})
	assert.Equal(t, p.Len(), 1)
	assert.NotNil(t, p)
	p.Start()

	j1 := p.CreateAndSubmitJob("input")
	assert.NotNil(t, j1)
	j2 := p.CreateAndSubmitJob("more stuff")
	assert.NotNil(t, j2)

	o1 := p.Result()
	o2 := p.Result()

	fmt.Printf("o1: %#v\n", o1)
	fmt.Printf("o2: %#v\n", o2)

	p.Stop()

	for _, jid := range jobsList {
		assert.Equal(t, jobsRan[jid], 1)
	}
}

func TestWorkerPoolWithTiming(t *testing.T) {
	p := NewWokerPool(1, func(j *Job) (interface{}, error) {
		fmt.Printf("Job: %#v\n", j)
		return "here", nil
	})
	assert.NotNil(t, p)

	called := int64(0)
	p.TimingLogger = func(jobID int64, inQueue, inWorker time.Duration) {
		fmt.Printf("jobID: %X time spent in queue: %s\n", jobID, inQueue)
		fmt.Printf("jobID: %X time spent in worker: %s\n", jobID, inWorker)
		atomic.AddInt64(&called, 1)
	}
	p.Start()

	j1 := p.CreateAndSubmitJob("input")
	assert.NotNil(t, j1)
	j2 := p.CreateAndSubmitJob("more stuff")
	assert.NotNil(t, j2)

	o1 := p.Result()
	o2 := p.Result()

	fmt.Printf("o1: %#v\n", o1)
	fmt.Printf("o2: %#v\n", o2)

	p.Stop()
	assert.Equal(t, called, int64(2))
}

func TestWorkerPoolDefaultSize(t *testing.T) {
	fmt.Printf("DefaultWorkerPoolSize: %d\n", DefaultWorkerPoolSize)
	assert.True(t, DefaultWorkerPoolSize > 0)
}

func TestWorkerPoolError(t *testing.T) {
	lock := sync.Mutex{}
	tofail := true
	jobsList := []int64{}
	jobsRan := map[int64]int{}
	p := NewWokerPool(1, func(j *Job) (interface{}, error) {
		fmt.Printf("Job: %#v\n", j)
		/////
		lock.Lock()
		defer lock.Unlock()

		if v, ok := jobsRan[j.ID]; ok {
			jobsRan[j.ID] = v + 1
		} else {
			jobsRan[j.ID] = 1
		}
		jobsList = append(jobsList, j.ID)
		if tofail {
			tofail = false
			return nil, fmt.Errorf("Something broke")
		}
		/////
		return "here", nil
	})
	assert.NotNil(t, p)
	p.Start()

	j1 := p.CreateAndSubmitJob("input")
	assert.NotNil(t, j1)
	j2 := p.CreateAndSubmitJob("more stuff")
	assert.NotNil(t, j2)

	o1 := p.Result()
	o2 := p.Result()

	if o1.Err != nil {
		assert.Nil(t, o1.Data)
		assert.NotNil(t, o1.Err)
		assert.Equal(t, o2.Data.(string), "here")
	} else {
		assert.Equal(t, o1.Data.(string), "here")
		assert.Nil(t, o2.Data)
		assert.NotNil(t, o2.Err)
	}

	fmt.Printf("o1: %#v\n", o1)
	fmt.Printf("o2: %#v\n", o2)

	p.Stop()
	assert.False(t, p.HasResult())

	assert.Equal(t, len(jobsList), 2)
	for _, jid := range jobsList {
		assert.Equal(t, jobsRan[jid], 1)
	}
}
