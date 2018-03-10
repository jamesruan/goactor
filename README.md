# goactor
actor model in go.

## usage/example
```
package main

import "github.com/jamesruan/goactor"
import "time"
import "fmt"

type Any = interface{}

func main() {
	pid := goactor.Spawn()
	go func() {
		matchAny := func(in Any) (bool, Any) {
			return true, in
		}
		handleTimeout := func(p *goactor.Pid) Any {
			return "exiting"
		}

		defer pid.Exit(goactor.NormalExit(""))

		for {
			v := pid.ReceiveWithTimeout(matchAny, 500*time.Millisecond, handleTimeout)
			fmt.Printf("%v got %v\n", pid, v)
			if v == "exiting" {
				return
			}
		}
	}()

	goactor.Send(pid, "test")
	goactor.Send(pid, "test timeout before")
	time.Sleep(1 * time.Second)
	goactor.Send(pid, "test timeout after")
	goactor.WaitAll() //make sure no actor is waiting
}
//<60.038µs> got test
//<60.038µs> got test timeout before
//<60.038µs> got exiting
```
