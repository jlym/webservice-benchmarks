package util

import "sync"

type StopSender struct {
	shouldStopC chan struct{}
	wg          *sync.WaitGroup
}

func NewStopSender() *StopSender {
	return &StopSender{
		shouldStopC: make(chan struct{}),
		wg:          &sync.WaitGroup{},
	}
}

func (ss *StopSender) StopAndWait() {
	close(ss.shouldStopC)
	ss.wg.Wait()
}

func (ss *StopSender) NewReciever() *StopReciever {
	ss.wg.Add(1)
	return &StopReciever{
		ShouldStopC: ss.shouldStopC,
		wg:          ss.wg,
	}
}

type StopReciever struct {
	ShouldStopC <-chan struct{}
	wg          *sync.WaitGroup
}

func (sr *StopReciever) Done() {
	sr.wg.Done()
}

func (sr *StopReciever) ShouldStop() bool {
	select {
	case <-sr.ShouldStopC:
		return true
	default:
		return false
	}
}

func (sr *StopReciever) ShouldContinue() bool {
	return !sr.ShouldStop()
}
