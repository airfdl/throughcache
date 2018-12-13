package throughcache

import (
	"errors"
	"time"
)

type Handler func()

type AsyncQue struct {
	JobQueue chan Handler
	Exit     chan struct{}
	Cap      int
	WorkCnt  int
}

func NewAsyncQue(cap int, workCnt int) *AsyncQue {
	asyncQue := &AsyncQue{
		JobQueue: make(chan Handler, cap),
		Exit:     make(chan struct{}),
		Cap:      cap,
		WorkCnt:  workCnt,
	}
	asyncQue.run()
	return asyncQue
}

func (s *AsyncQue) run() {
	for i := 0; i < s.WorkCnt; i++ {
		go s.Worker()
	}
}

func (s *AsyncQue) Name() string {
	return "AsyncQue"
}

func (s *AsyncQue) Send(h Handler) error {
	select {
	case <-s.Exit:
		return errors.New("sync queue was closed")
	case s.JobQueue <- h:
		return nil
	default:
		return errors.New("sync queue was full")
	}
}

func (s *AsyncQue) Worker() {
	for {
		s.worker()
		time.Sleep(time.Second)
		//check this Queue is closed
		select {
		case <-s.Exit:
			return
		default:
			logs.Error("unexpect:%v", "worker will be restart now")
		}
	}
}

func (s *AsyncQue) worker() {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("unexpect error in async queue: %v", r)
		}
	}()
	for {
		select {
		case f := <-s.JobQueue:
			f()
		case <-s.Exit:
			return
		}
	}
}

func (s *AsyncQue) Close() {
	close(s.Exit)
}
