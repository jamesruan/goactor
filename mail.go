package goactor

import "sync"

type Mail struct {
	t mailType
	p Any
}

type mailType int

const (
	mailUser mailType = iota
	mailSysExit
)

var mail_pool *sync.Pool

func init() {
	mail_pool = &sync.Pool{
		New: func() Any {
			return Mail{}
		},
	}
}
