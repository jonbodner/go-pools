package workers

import (
	"crypto/rand"
	"encoding/binary"
	"runtime"
	"sync"
	"time"
)

type WorkerFunc func(job *Job) (interface{}, error)

type TimingFunc func(jobID int64, inQueue, inWorker time.Duration)

type Job struct {
	ID            int64
	inTime        int64 // inserted timestamp
	inWorkerTime  int64
	outWorkerTime int64
	Data          interface{}
	//failureCount  uint32
}

type Result struct {
	JobID int64
	Data  interface{}
	Err   error
}

type WorkerPool struct {
	size         uint32
	workerFunc   WorkerFunc
	in           chan *Job
	out          chan *Result
	doneWg       sync.WaitGroup
	TimingLogger TimingFunc
}

var (
	DefaultWorkerPoolSize = runtime.NumCPU()
)

func NewWokerPool(size uint32, workerFunc WorkerFunc) *WorkerPool {
	if workerFunc == nil {
		return nil
	}
	return &WorkerPool{size: size, workerFunc: workerFunc,
		in: make(chan *Job, size), out: make(chan *Result, size*4)}
}

func (w *WorkerPool) Start() {
	for i := uint32(0); i < w.size; i++ {
		w.doneWg.Add(1)
		go func() {
		WorkerLoop:
			for {
				select {
				case job, ok := <-w.in:
					if job != nil {
						job.inWorkerTime = timestamp()
						result, err := w.workerFunc(job)
						job.outWorkerTime = timestamp()
						w.out <- &Result{Data: result, JobID: job.ID, Err: err}
						if w.TimingLogger != nil {
							w.TimingLogger(job.ID, time.Duration(job.inWorkerTime-job.inTime)*time.Nanosecond,
								time.Duration(job.outWorkerTime-job.inWorkerTime)*time.Nanosecond)
						}
					}
					if !ok {
						break WorkerLoop
					}
				default:
				}
			}
			w.doneWg.Done()
		}()
	}
}

func (w *WorkerPool) CreateAndSubmitJob(data interface{}) *Job {
	job := w.CreateJob(data)
	w.SubmitJob(job)
	return job
}

// CreateJob will build a job that needs to be submitted
func (w *WorkerPool) CreateJob(data interface{}) *Job {
	job := &Job{ID: makeID(), Data: data, inTime: timestamp()}
	return job
}

// SubmitJob will add a job to the worker queue, but if there
// is no room available, it will block.
func (w *WorkerPool) SubmitJob(job *Job) {
	w.in <- job
}

// Result will return the result from the worker, but if no result
// is available, then it will block
func (w *WorkerPool) Result() *Result {
	return <-w.out
}

func (w *WorkerPool) HasResult() bool {
	return len(w.out) > 0
}

func (w *WorkerPool) OutChannel() chan *Result {
	return w.out
}

func (w *WorkerPool) Stop() {
	close(w.in)
	w.doneWg.Wait()
}

func (w *WorkerPool) Len() int {
	return int(w.size)
}

func makeID() int64 {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return -1
	}
	result, _ := binary.Varint(b)
	return result
}

func timestamp() int64 {
	return time.Now().UTC().UnixNano()
}
