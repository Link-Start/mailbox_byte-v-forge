package hotstream

func (s *Subscription) Close() {
	if s == nil || s.inner == nil {
		return
	}
	if s.hub != nil {
		s.hub.unsubscribe(s.inner, nil)
		return
	}
	s.inner.close(nil)
}

func (s *Subscription) Err() error {
	if s == nil || s.inner == nil {
		return nil
	}
	return s.inner.err
}

func (s *Subscription) Done() <-chan struct{} {
	if s == nil || s.inner == nil {
		closed := make(chan struct{})
		close(closed)
		return closed
	}
	return s.inner.done
}

func (s *subscription) close(err error) {
	s.once.Do(func() {
		s.err = err
		close(s.done)
		close(s.events)
	})
}
