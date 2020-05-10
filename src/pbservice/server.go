package pbservice

import "net"
import "fmt"
import "net/rpc"
import "log"
import "time"
import "viewservice"
import "sync"
import "sync/atomic"
import "os"
import "syscall"
import "math/rand"



type PBServer struct {
	mu         sync.Mutex
	l          net.Listener
	dead       int32 // for testing
	unreliable int32 // for testing
	me         string
	vs         *viewservice.Clerk
	// Your declarations here.
	viewnum	   int32
	kv	       map[string]string
	status	   int32 // 1 primary, 2 backup, 0 offline
	primary    string // name of primary server (if not this one)
	backup     string // name of backup
}


func (pb *PBServer) Get(args *GetArgs, reply *GetReply) error {

	// Your code here.
	if (status == 1) {
		//do primary work
		reply.Value, err = kv[args.Key]
		if (err != nil) {
			reply.Err = ErrNoKey
		}
		args.Caller = me
		call(backup, "PBServer.Get", args, reply)
	}
	else if (status == 2) {
		//do backup work
		if (args.Caller == primary) {
			//primary has forwarded call to us, verify we have same value for that key
			reply.Value, err = kv[args.Key]
			if (err != nil) {
				reply.Err = ErrNoKey
			}
			else {
				reply.Err = OK
			}
		}
		else {
			// only handle get if its forwarded from primary
			reply.Err = ErrWrongServer
		}	
	}

	return nil
}


func (pb *PBServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) error {

	// Your code here.
	if (status == 1) {
		//do primary work
		if (args.Put == "Put") {
			kv[args.Key] = args.Value
		}
		else {
			kv[args.Key] += args.Value
		}
		args.Caller = me
		call(backup, "PBServer.PutAppend", args, reply)
	}
	else if (status == 2) {
		//do backup work
		if (args.Caller == primary) {
			if (args.Put == "Put") {
				kv[args.Key] = args.Value
			}
			else {
				kv[args.Key] += args.Value
			}
			reply.Err = OK
		}
		else {
			reply.Err = ErrWrongServer
		}
	}

	return nil
}


//
// ping the viewserver periodically.
// if view changed:
//   transition to new view.
//   manage transfer of state from primary to new backup.
//
func (pb *PBServer) tick() {

	// ping view service
	current_view := pb.vs.Ping(pb.viewnum) // need to somehow increment this and make it related to viewnum of viewserver
	pb.viewnum = current_view.Viewnum
	primary = current_view.Primary
	backup = current_view.Backup
	// check if this server is primary or backup
	// this will influence what kind of puts/gets it accepts and how to respond....?
	if (me == primary) {
		//we are primary server
		status = 1
	}
	else if (me == backup) {
		status = 2
	}
	else {
		status = 0
	}

}

// tell the server to shut itself down.
// please do not change these two functions.
func (pb *PBServer) kill() {
	atomic.StoreInt32(&pb.dead, 1)
	pb.l.Close()
}

// call this to find out if the server is dead.
func (pb *PBServer) isdead() bool {
	return atomic.LoadInt32(&pb.dead) != 0
}

// please do not change these two functions.
func (pb *PBServer) setunreliable(what bool) {
	if what {
		atomic.StoreInt32(&pb.unreliable, 1)
	} else {
		atomic.StoreInt32(&pb.unreliable, 0)
	}
}

func (pb *PBServer) isunreliable() bool {
	return atomic.LoadInt32(&pb.unreliable) != 0
}


func StartServer(vshost string, me string) *PBServer {
	pb := new(PBServer)
	pb.me = me
	pb.vs = viewservice.MakeClerk(me, vshost)
	// Your pb.* initializations here.
	pb.viewnum = 0

	rpcs := rpc.NewServer()
	rpcs.Register(pb)

	os.Remove(pb.me)
	l, e := net.Listen("unix", pb.me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	pb.l = l

	// please do not change any of the following code,
	// or do anything to subvert it.

	go func() {
		for pb.isdead() == false {
			conn, err := pb.l.Accept()
			if err == nil && pb.isdead() == false {
				if pb.isunreliable() && (rand.Int63()%1000) < 100 {
					// discard the request.
					conn.Close()
				} else if pb.isunreliable() && (rand.Int63()%1000) < 200 {
					// process the request but force discard of reply.
					c1 := conn.(*net.UnixConn)
					f, _ := c1.File()
					err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
					if err != nil {
						fmt.Printf("shutdown: %v\n", err)
					}
					go rpcs.ServeConn(conn)
				} else {
					go rpcs.ServeConn(conn)
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && pb.isdead() == false {
				fmt.Printf("PBServer(%v) accept: %v\n", me, err.Error())
				pb.kill()
			}
		}
	}()

	go func() {
		for pb.isdead() == false {
			pb.tick()
			time.Sleep(viewservice.PingInterval)
		}
	}()

	return pb
}