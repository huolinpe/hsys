package gredis

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
	"unsafe"
)

type Cmder interface {
	Name() string
	FullName() string
	Args() []interface{}
	String() string
	stringArg(int) string
	firstKeyPos() int8
	SetFirstKeyPos(int8)

	readTimeout() *time.Duration
	readReply(rd *bufio.Reader) error

	SetErr(error)
	Err() error
}

func setCmdsErr(cmds []Cmder, e error) {
	for _, cmd := range cmds {
		if cmd.Err() == nil {
			cmd.SetErr(e)
		}
	}
}

type baseCmd struct {
	ctx    context.Context
	args   []interface{}
	err    error
	keyPos int8

	_readTimeout *time.Duration
}

// var _ Cmder = (*Cmd)(nil)

func (cmd *baseCmd) Name() string {
	if len(cmd.args) == 0 {
		return ""
	}
	// Cmd name must be lower cased.
	// return internal.ToLower(cmd.stringArg(0))
	return ""
}

func (cmd *baseCmd) FullName() string {
	switch name := cmd.Name(); name {
	case "cluster", "command":
		if len(cmd.args) == 1 {
			return name
		}
		if s2, ok := cmd.args[1].(string); ok {
			return name + " " + s2
		}
		return name
	default:
		return name
	}
}

func (cmd *baseCmd) Args() []interface{} {
	return cmd.args
}

func (cmd *baseCmd) stringArg(pos int) string {
	if pos < 0 || pos >= len(cmd.args) {
		return ""
	}
	arg := cmd.args[pos]
	switch v := arg.(type) {
	case string:
		return v
	default:
		// TODO: consider using appendArg
		return fmt.Sprint(v)
	}
}

func (cmd *baseCmd) firstKeyPos() int8 {
	return cmd.keyPos
}

func (cmd *baseCmd) SetFirstKeyPos(keyPos int8) {
	cmd.keyPos = keyPos
}

func (cmd *baseCmd) SetErr(e error) {
	cmd.err = e
}

func (cmd *baseCmd) Err() error {
	return cmd.err
}

func (cmd *baseCmd) readTimeout() *time.Duration {
	return cmd._readTimeout
}

func (cmd *baseCmd) setReadTimeout(d time.Duration) {
	cmd._readTimeout = &d
}

type StringCmd struct {
	baseCmd
	val string
}

// var _ Cmder = (*StringCmd)(nil)

func NewStringCmd(ctx context.Context, args ...interface{}) *StringCmd {
	return &StringCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *StringCmd) SetVal(val string) {
	cmd.val = val
}

func (cmd *StringCmd) Val() string {
	return cmd.val
}

func (cmd *StringCmd) Result() (string, error) {
	return cmd.Val(), cmd.err
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func (cmd *StringCmd) Bytes() ([]byte, error) {
	return StringToBytes(cmd.val), cmd.err
}

func (cmd *StringCmd) Bool() (bool, error) {
	if cmd.err != nil {
		return false, cmd.err
	}
	return strconv.ParseBool(cmd.val)
}

func (cmd *StringCmd) Int() (int, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.Atoi(cmd.Val())
}

func (cmd *StringCmd) Int64() (int64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseInt(cmd.Val(), 10, 64)
}

func (cmd *StringCmd) Uint64() (uint64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseUint(cmd.Val(), 10, 64)
}

func (cmd *StringCmd) Float32() (float32, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	f, err := strconv.ParseFloat(cmd.Val(), 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func (cmd *StringCmd) Float64() (float64, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	return strconv.ParseFloat(cmd.Val(), 64)
}

func (cmd *StringCmd) Time() (time.Time, error) {
	if cmd.err != nil {
		return time.Time{}, cmd.err
	}
	return time.Parse(time.RFC3339Nano, cmd.Val())
}

func (cmd *StringCmd) Scan(val interface{}) error {
	if cmd.err != nil {
		return cmd.err
	}
	// return proto.Scan([]byte(cmd.val), val)
	return nil
}

func (cmd *StringCmd) String() string {
	// return cmdString(cmd, cmd.val)
	return ""
}

func (cmd *StringCmd) readReply(rd *bufio.Reader) (err error) {
	// cmd.val, err = rd.ReadString()
	log.Println("enter readReply")
	b, err := rd.ReadSlice('\n')
	if err != nil {
		return err
	}
	if len(b) <= 2 || b[len(b)-1] != '\n' || b[len(b)-2] != '\r' {
		return fmt.Errorf("redis: invalid reply: %q", b)
	}

	log.Printf("StringCmd slice:%s", string(b))

	n, err := strconv.Atoi(BytesToString(b[1 : len(b)-2]))
	if err != nil {
		log.Println("StringCmd err return")
		return err
	}

	log.Printf("StringCmd n:%d", n)

	bt := make([]byte, n+2)
	_, err = io.ReadFull(rd, bt)
	if err != nil {
		return err
	}

	// log.Printf("StringCmd bt:%s", string(bt))

	cmd.val = BytesToString(bt[:n])
	return nil
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//----------------------------------------------------------------------

type StatusCmd struct {
	baseCmd
	val string
}

// var _ Cmder = (*StatusCmd)(nil)

func NewStatusCmd(ctx context.Context, args ...interface{}) *StatusCmd {
	return &StatusCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *StatusCmd) SetVal(val string) {
	cmd.val = val
}

func (cmd *StatusCmd) Val() string {
	return cmd.val
}

func (cmd *StatusCmd) Result() (string, error) {
	return cmd.val, cmd.err
}

func (cmd *StatusCmd) String() string {
	// return cmdString(cmd, cmd.val)
	return ""
}

func (cmd *StatusCmd) readReply(rd *bufio.Reader) (err error) {
	// cmd.val, err = rd.ReadString()
	log.Println("enter readReply")
	b, err := rd.ReadSlice('\n')
	if err != nil {
		// if err != bufio.ErrBufferFull {
		// 	return nil, err
		// }

		// full := make([]byte, len(b))
		// copy(full, b)

		// b, err = r.rd.ReadBytes('\n')
		// if err != nil {
		// 	return nil, err
		// }

		// full = append(full, b...) //nolint:makezero
		// b = full
		return err
	}
	if len(b) <= 2 || b[len(b)-1] != '\n' || b[len(b)-2] != '\r' {
		return fmt.Errorf("redis: invalid reply: %q", b)
	}

	log.Printf("StatusCmd slice:%s", string(b))

	cmd.val = string(b[:len(b)-2])
	return nil
}

func setStringBytes(cmd Cmder) []byte {
	var wrbytes []byte
	wrbytes = append(wrbytes, '*')
	args := cmd.Args()
	wrbytes = strconv.AppendUint(wrbytes, uint64(len(args)), 10)
	wrbytes = append(wrbytes, '\r', '\n')

	for _, arg := range args {
		wrbytes = append(wrbytes, '$')

		wrbytes = strconv.AppendUint(wrbytes, uint64(len(arg.(string))), 10)
		wrbytes = append(wrbytes, '\r', '\n')
		// byarg := StringToBytes(arg.(string))
		argstring := StringToBytes(arg.(string))
		wrbytes = append(wrbytes, argstring...)
		wrbytes = append(wrbytes, '\r', '\n')
	}

	log.Println(string(wrbytes))
	return wrbytes
}

func writeCmd(wr *bufio.Writer, cmd Cmder) error {
	wrBytes := setStringBytes(cmd)
	if _, err := wr.Write(wrBytes); err != nil {
		return err
	}
	return nil
}
