package main

type mailboxRuntimeClosers []func()

func (c *mailboxRuntimeClosers) add(close func()) {
	if close != nil {
		*c = append(*c, close)
	}
}

func (c mailboxRuntimeClosers) close() {
	for i := len(c) - 1; i >= 0; i-- {
		c[i]()
	}
}
