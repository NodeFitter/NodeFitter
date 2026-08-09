package controller

import (
	"net"
	"net/rpc"
	"os"

	"github.com/NodeFitter/NodeFitter/controller/abstraction"
)

func Serve(socket string, ctrl abstraction.Icontroller) error {
	_ = os.Remove(socket)

	l, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}

	if err := rpc.Register(ctrl); err != nil {
		_ = l.Close()
		return err
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			go rpc.ServeConn(conn)
		}
	}()

	return nil
}
