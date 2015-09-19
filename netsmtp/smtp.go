package netsmtp

// import (
// 	"bitbucket.org/chrj/smtpd"
// 	"github.com/coreos/go-etcd/etcd"
// )
//
// func smtpSetup(cfg *Config, etc *etcd.Client) chan error {
// 	// FIXME: allow cfg to enable/disable this service
//
// 	smtp := &smtpd.Server{Handler: smtpHandler}
// 	exit := make(chan error)
//
// 	go func() {
// 		exit <- smtp.ListenAndServe("0.0.0.0:25") // TODO: should use cfg to define the listening ip/port
// 	}()
//
// 	return exit
// }
//
// func smtpHandler(peer smtpd.Peer, env smtpd.Envelope) error {
// 	// FIXME: make this do something.  right now it just discards what it gets.
//
// 	return nil
// }
