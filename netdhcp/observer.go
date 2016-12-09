package netdhcp

/*
func ManageConfigTest() {
	var gobs GlobalObserver
	var nobs NetworkObserver
	var iobs InstanceObserver
	defer gobs.Close()
	defer nobs.Close()
	defer iobs.Close()
	gpulse := gobs.Subscribe()
	npulse := nobs.Subscribe()
	ipulse := iobs.Subscribe()
	global, err := gobs.Read()
	network, err := nobs.Read()
	instance, err := iobs.Read()
	for {
		select {
			<-gpulse:
				global, err = gobs.Read()
			<-npulse:
				network, err = nobs.Read()
			<-ipulse:
				instance, err = iobs.Read()
		}
	}
}
*/
