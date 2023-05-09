package hrpc

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"reflect"
	"sync"
	"time"
	// "fmt"
)

type RPCHandler func(param ...interface{}) interface{}

//var clientFilters = make(map[string]RPCHandler)

type InvokeData struct {
	Token     int64  //单次调用序列号
	Method    string //调用方法名
	Done      bool   //是否调用返回
	Err       string //错误信息
	ParamData []byte //参数
}

type RemoteStream struct {
	conn      net.Conn
	writer    *bufio.Writer
	reader    *bufio.Reader
	encoder   *gob.Encoder
	decoder   *gob.Decoder
	mutex     *sync.Mutex
	nextToken int64
	waits     sync.Map
	timeout   time.Duration
	Handlers  map[string]RPCHandler //rpc处理函数,注册不同的处理函数
}

func (s *RemoteStream) InitConn(conn net.Conn, timeout int, rsize int, wsize int) {
	s.conn = conn
	s.reader = bufio.NewReaderSize(conn, rsize)
	s.writer = bufio.NewWriterSize(conn, wsize)
	s.decoder = gob.NewDecoder(s.reader)
	s.encoder = gob.NewEncoder(s.writer)
	s.timeout = time.Second * time.Duration(timeout)
	s.mutex = new(sync.Mutex)
	s.Handlers = make(map[string]RPCHandler) //处理函数注册 todo
}

func (s *RemoteStream) Call(params ...interface{}) error {
	var methodPos int
	for i, v := range params {
		k := reflect.TypeOf(v).Kind()
		if k == reflect.String {
			methodPos = i
			break
		} else if k != reflect.Ptr {
			return errors.New("return args should be point")
		}
	}

	stm := &bytes.Buffer{}
	enc := gob.NewEncoder(stm)
	for i := methodPos; i < len(params); i++ {
		err := enc.Encode(params[i])
		if err != nil {
			return errors.New("endcode params err")
		}
	}

	//构造调用信息
	var ivk InvokeData
	ivk.Token = s.nextToken
	ivk.Method = params[methodPos].(string)
	ivk.Done = false
	ivk.ParamData = stm.Bytes()

	ch := make(chan InvokeData, 1)

	s.mutex.Lock()
	s.nextToken++
	s.waits.Store(ivk.Token, ch)
	s.send(&ivk)
	s.mutex.Unlock()

	select {
	case ivk = <-ch:
		break
	case <-time.After(s.timeout):
		return errors.New("recv timeout")
	}

	if ivk.Err != "" {
		return errors.New(ivk.Err)
	}

	// log.Printf("ivk.paramdata return: %d, %s\n", methodPos, ivk.ParamData)
	// stm = bytes.NewBuffer(ivk.ParamData)
	// dec := gob.NewDecoder(stm)
	// for i := 0; i < methodPos; i++ {
	// 	err := dec.Decode(params[i])
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	*params[0].(*string) = string(ivk.ParamData[:])
	return nil
}

func (s *RemoteStream) send(ivk *InvokeData) error {
	s.conn.SetWriteDeadline(time.Now().Add(s.timeout))
	err := s.encoder.Encode(ivk)
	if err == nil {
		err = s.writer.Flush()
		if err == nil {
			return nil
		}
	}
	return err
}

func (s *RemoteStream) Run() {
	for {
		s.conn.SetReadDeadline(time.Now().Add(s.timeout))
		var ivk InvokeData
		err := s.decoder.Decode(&ivk)
		if err != nil {
			break
		}
		go s.invoke(&ivk)
	}
	s.conn.Close()
}

func (s *RemoteStream) invoke(ivk *InvokeData) {
	if ivk.Done { //回调
		ch, ok := s.waits.Load(ivk.Token)
		if !ok {
			return
		}
		s.waits.Delete(ivk.Token)
		ch.(chan InvokeData) <- *ivk

	} else { //远端调用
		err := s.apply(ivk)
		if err != nil {
			ivk.Done = true
			ivk.ParamData = nil
			ivk.Err = err.Error()
			s.mutex.Lock()
			s.send(ivk)
			s.mutex.Unlock()
		}
	}
}

func (s *RemoteStream) apply(ivk *InvokeData) error {
	// v := reflect.ValueOf(s.handler)
	// method := v.MethodByName(ivk.Method)
	// if !method.IsValid() {
	// 	return errors.New("method not found")
	// }

	// mt := method.Type()
	// in := []reflect.Value{}
	// stm := bytes.NewBuffer(ivk.Data)
	// dec := gob.NewDecoder(stm)
	// for i := 0; i < mt.NumIn(); i++ {
	// 	arg := reflect.New(mt.In(i))
	// 	err := dec.DecodeValue(arg)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to decode args %d", i)
	// 	}
	// 	in = append(in, arg.Elem())
	// }

	// out := method.Call(in)

	// callfunc, err := s.Handlers[ivk.Method]
	// if !err {
	// 	return errors.New("don't find this method")
	// }

	out := GetRemoteIp()

	// 返回值打包
	// stm := &bytes.Buffer{}
	// enc := gob.NewEncoder(stm)
	// err = enc.EncodeValue(string(out))
	// if err {
	// 	return fmt.Errorf("failed to encode out")
	// }
	// for i := 0; i < len(out); i++ {
	// 	err := enc.EncodeValue(out[i])
	// 	if err != nil {
	// 		return fmt.Errorf("failed to encode out %d", i)
	// 	}
	// }

	ivk.Done = true
	ivk.ParamData = []byte(out)
	ivk.Err = ""
	s.mutex.Lock()
	s.send(ivk) // send失败不要返回err,因为上层也没法再发送给对端了
	s.mutex.Unlock()
	return nil

}

//rpc函数
func GetRemoteIp() string {
	log.Println("rpc call come here!")
	return "call rpc sucess"
}

// func (s* RemoteStream)RegisterFunc() err {
// 	return nil
// }
