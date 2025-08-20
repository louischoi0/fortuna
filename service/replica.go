package service

import (
	"fortuna/core/model"
	"fortuna/rpc"
	"fortuna/swift"
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

	swift		*swift.TCPServer
}

func NewReplica() *Replica {
	swift := swift.NewServer()
	return &Replica{
		swift: swift,
		Universe: make(map[string]*model.Space),
		connected: false,
	}
}

func (rp *Replica) Connect(addr string) error {
	conn, err := rpc.Connect(addr)

	if err != nil {
		return err
	}

	log.Println("successfully connected")
	rp.conn = conn
	rp.connected = true

	return nil
}

func (rp *Replica) Run() error {
	if !rp.connected {
		log.Fatalf("replica not connected")
	}

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
			log.Println("ping")
			if _, err := rp.conn.Write(; err != nil {
				log.Fatalf(err.Error())
				return err
			}
		}
	}
}



