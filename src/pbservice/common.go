package pbservice

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongServer = "ErrWrongServer"
	ErrRepeatCall  = "ErrRepeatCall"
)

type Err string

// Put or Append
type PutAppendArgs struct {
	Key    string
	Value  string
	Caller string
	// You'll have to add definitions here.
	Put string
	ID  int64
	// Field names must start with capital letters,
	// otherwise RPC will break.
}

type PutAppendReply struct {
	Err Err
}

type GetArgs struct {
	Key          string
	PrimaryValue string
	// You'll have to add definitions here.
	Caller string // id of server making call
	// something to make get calls unique
	ID int64
}

type GetReply struct {
	Err   Err
	Value string
}

type TransferKVArgs struct {
	CallIDs map[int64]bool
	KV      map[string]string
}

type TransferReply struct {
	Err Err
}

// Your RPC definitions here.
/*
type GetForward struct {
	Key          string
	PrimaryValue string
	Caller       string // id of server making call
	ID           int64
}

type PutAppendForward struct {
	Key    string
	Value  string
	Caller string
	// You'll have to add definitions here.
	Put string
	ID  int64
	// Field names must start with capital letters,
	// otherwise RPC will break.
}
*/
