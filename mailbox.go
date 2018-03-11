package goactor

type listMailbox struct {
	queue
}

// MailMatcher checks the mail in mailbox, returns true will remove the mail from the box.
//
// Note: MailMatcher must not block
type MailMatcher func(in Any) (matched bool)

func newListMailbox() *listMailbox {
	return &listMailbox{
		makeQueue(),
	}
}

func (m *listMailbox) enqueue(v Any) {
	m.enqueueType(v, mailUser)
}

func (m *listMailbox) enqueueType(v Any, t mailType) {
	mail := mail_pool.Get().(Mail)
	mail.t = t
	mail.p = v
	m.queue.enqueue(mail)
}

func (m *listMailbox) peek() (v Any, ok bool) {
	return m.peekType(mailUser)
}

func (m *listMailbox) peekType(t mailType) (v Any, ok bool) {
	if v, ok = m.queue.peek(); ok {
		mail := v.(Mail)
		ok = mail.t == t
		if ok && t == mailUser {
			v = mail.p
		}
	}
	return
}

func (m *listMailbox) dequeue() (ok bool) {
	v, ok := m.queue.dequeue()
	if ok {
		mail_pool.Put(v)
	}
	return
}

func (m *listMailbox) transferFront(from *listMailbox) {
	m.queue.transferFront(from.queue)
}
