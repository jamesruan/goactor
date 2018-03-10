package goactor

import "container/list"

type listMailbox struct {
	*list.List
}

type Any = interface{}

// MailMatcher checks the mail in mailbox, returns true will remove the mail from the box.
//
// Note: MailMatcher must not block
type MailMatcher func(in Any) (matched bool)

func newListMailbox() *listMailbox {
	return &listMailbox{
		list.New(),
	}
}

func (m *listMailbox) enqueue(v Any) {
	m.PushBack(v)
}

func (m *listMailbox) peek() (v Any, ok bool) {
	front := m.Front()
	ok = front != nil
	if ok {
		v = front.Value
	}
	return
}

func (m *listMailbox) dequeue() (v Any, ok bool) {
	front := m.Front()
	ok = front != nil
	if ok {
		v = m.Remove(front)
	}
	return
}

func (m *listMailbox) isEmpty() bool {
	return m.Front() == nil
}

func (m *listMailbox) unshiftAll(from *listMailbox) {
	if !from.isEmpty() {
		m.PushFrontList(from.List)
		from.Init()
	}
}
