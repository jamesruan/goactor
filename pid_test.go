package goactor

import "testing"
import "time"

import "net/http"
import _ "net/http/pprof"

func init() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}

func TestTimeout(t *testing.T) {
	pid := Spawn()
	go func() {
		defer pid.Exit(NormalExit(""))
		for {
			v := pid.ReceiveWithTimeout(func(in Any) bool {
				if in == "test" {
					t.Logf("%v consuming %#v", pid, in)
					return true
				} else {
					t.Logf("%v ignoring %#v", pid, in)
					return false
				}
			}, 10*time.Millisecond, func(p *Pid) Any {
				t.Logf("%v timeout", pid)
				return "exit"
			})
			t.Logf("%v got %#v", pid, v)
			if v == "exit" {
				return
			}
		}
	}()
	Send(pid, "test ignore 1")
	Send(pid, "test ignore 2")
	Send(pid, "test")
	Send(pid, "test ignore 3")
	Send(pid, "test ignore 4")
	time.Sleep(50 * time.Millisecond)
	Send(pid, "exit")
	WaitAll()
}

func TestSleep(t *testing.T) {
	pid := Spawn()
	go func() {
		var none MailMatcher
		v := pid.ReceiveWithTimeout(none, 100*time.Millisecond, func(p *Pid) Any {
			return "timeout"
		})
		t.Logf("%v got %#v", pid, v)
		pid.Exit(NormalExit(""))
	}()
	WaitAll()
}

func TestTimeoutZero(t *testing.T) {
	pid := Spawn()
	go func() {
		v := pid.ReceiveWithTimeout(func(in Any) bool {
			t.Logf("%v matching %v", pid, in)
			if in == "key" {
				return true
			} else {
				return false
			}
		}, 0, func(p *Pid) Any {
			r := p.Receive(func(in Any) bool {
				return true
			})
			t.Logf("%v receive %#v", pid, r)
			return r
		})
		t.Logf("%v matched %v", pid, v)
		pid.Exit(NormalExit(""))
	}()
	Send(pid, "test")
	Send(pid, "key")
	WaitAll()
}
