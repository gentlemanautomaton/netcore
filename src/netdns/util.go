package netdns

func makeOnce() chan struct{} {
	once := make(chan struct{}, 1)
	once <- struct{}{}
	close(once)
	return once
}
