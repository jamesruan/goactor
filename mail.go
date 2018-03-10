package goactor

type SysMail struct {
	t sysMailType
	p Any
}

type sysMailType int

const sysMailExit = iota
