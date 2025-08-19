package service

import (
	"fortuna/core/model"
	"fortuna/rpc"
	"sync"
	"github.com/linxGnu/grocksdb"
	"fmt"
	"log"
	"net"
	"bufio"
	"time"
	"context"
	"os/signal"
	"os"
)

type Replica struct {
	mu		sync.Mutex	
	universeDB	*grocksdb.DB
	Universe	map[string]*model.Space

	conn		net.Conn
	connected	bool
	masterInfo	*nodeInfo
}

func NewReplica() *Replica {
	return &Replica{
		Universe: make(map[string]*model.Space),
		connected: false,
	}
}

type nodeInfo struct {
	addr	string
}

func (rp *Replica) Connect(n *nodeInfo) error {
	conn, err := rpc.Connect(n.addr)

	if err != nil {
		return err
	}

	rp.conn = conn
	rp.connected = true
	rp.masterInfo = n

	return nil
}

func (rp *Replica) Run() error {
	r := bufio.NewReader(rp.conn)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	
	done := make(chan error, 1)
	go func() {
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				done <- err
				return 
			}
			log.Println("recieved buffer: %s", line)
		}
	}()

	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case err := <- done:
			return err
		case <- ping.C:
			if _, err := fmt.Fprintln(rp.conn, "PING"); err != nil {
				return err
			}
		}
	}
}



