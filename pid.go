package goactor

import "time"
import "sync"

var pids_wg *sync.WaitGroup
var sysStartTime time.Time

func init() {
	pids_wg = new(sync.WaitGroup)
	sysStartTime = time.Now()
}

type Pid struct {
	timeoffset time.Duration
	incomming  *listMailbox
	mailbox    *listMailbox
	waiting    *listMailbox
	input      chan Any
	matching   chan Any
	inbox      chan Any
	system     chan Mail
	exited     chan struct{}
	*sync.Mutex
}

func Spawn() *Pid {
	pid := &Pid{
		timeoffset: time.Now().Sub(sysStartTime),
		incomming:  newListMailbox(),
		mailbox:    newListMailbox(),
		waiting:    newListMailbox(),
		input:      make(chan Any),
		matching:   make(chan Any, 1),
		inbox:      make(chan Any),
		system:     make(chan Mail),
		exited:     make(chan struct{}),
	}
	pids_wg.Add(1)
	go func() {
		defer pids_wg.Done()
		for {
			var head Any
			var inbox chan Any
			var ok bool
			if head, ok = pid.incomming.peek(); ok {
				inbox = pid.inbox
			}
			select {
			case sm := <-pid.system:
				if sm.t == mailSysExit {
					pid.kill()
					return
				}
			case m := <-pid.input:
				//TODO: filter
				pid.incomming.enqueue(m)
			case inbox <- head:
				pid.incomming.dequeue()
			}
		}
	}()

	return pid
}

func (p Pid) String() string {
	return "<" + p.timeoffset.String() + ">"
}

func WaitAll() {
	pids_wg.Wait()
}

func Send(dst *Pid, v Any) {
	go func() {
		select {
		case <-dst.exited:
		case dst.input <- v:
		}
	}()
}

func (p *Pid) Exit(reason ExitError) {
	p.system <- Mail{
		mailSysExit,
		reason,
	}
}

func (p *Pid) kill() {
	p.inbox = nil
	p.matching = nil
	p.mailbox.Init()
	p.waiting.Init()
	close(p.exited)
}

// Receive go through the internal mailbox to run MailMatcher on each mail in receiving order.
//
// If 'consume', return the mapped value; otherwise the mapped value will be ignored and Receive will block until the next comming mail and try again.
//
// Note: Receive or ReceiveWithTimeout must not be called simultaneously for each Pid.
func (p *Pid) Receive(m MailMatcher) (v Any) {
	return p.ReceiveWithTimeout(m, -1, nil)
}

// ReceiveWithTimeout works like Receive, but will block for at most duration 'd'.
//
// If 'consume', return the mapped value; otherwise the mapped will be ignored and it will continues wait for the next comming mail or timeout.
//
// If time is out, return the result of 'f' function.
//
// The internal mailbox will always be checked before time out.
//
// Note: Receive or ReceiveWithTimeout must not be called simultaneously for each Pid.
func (p *Pid) ReceiveWithTimeout(m MailMatcher, d time.Duration, f func(*Pid) Any) (v Any) {
	return p.doReceive(m, d, f)
}

func (p *Pid) doReceive(m MailMatcher, d time.Duration, f func(*Pid) Any) Any {
	for {
		var head Any
		var matching chan Any
		var timer *time.Timer
		var timeout <-chan time.Time
		head, ok := p.mailbox.peek()
		if ok {
			matching = p.matching
		} else {
			if d > 0 {
				timer = time.NewTimer(d)
				timeout = timer.C
			}
		}
		select {
		case <-p.exited:
			return nil
		case mail := <-p.inbox:
			p.mailbox.enqueue(mail)
		case matching <- head:
			p.mailbox.dequeue()
		case mail := <-p.matching:
			if m != nil {
				if m(mail) {
					tryStopTimer(timer)
					p.mailbox.transferFront(p.waiting)
					return mail
				}
			}
			p.waiting.enqueue(mail)
			if d == 0 {
				return f(p)
			}
		case <-timeout:
			return f(p)
		}
	}
}

func tryStopTimer(timer *time.Timer) {
	if timer != nil {
		// stop and drain the timer
		if !timer.Stop() {
			<-timer.C
		}
	}
}
