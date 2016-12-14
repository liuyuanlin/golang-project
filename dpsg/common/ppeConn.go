package common

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sync"
	//"time"
	"golang-project/dpsg/logger"
	"runtime/debug"
)

type methodType struct {
	method  reflect.Method
	ArgType reflect.Type
}

type cmdType struct {
	rtType reflect.Type
}

// 通讯包协议头，所有的头文件都要用这个
type cmdHead struct {
	MsgId  uint16 // 包ID，对应反序列化的json类型
	CmdLen uint16 // json数据段长度
}

type PpeConn struct {
	conn       net.Conn
	isShutDown bool
	shutDown   chan bool
	recvBuf    bytes.Buffer
	sendBuf    bytes.Buffer
	methods    map[string]*methodType
	msgs       map[uint16]*cmdType
	regitesr   interface{}
	lsend      sync.RWMutex
	send       chan bool
}

func CreatePpeConn(conn net.Conn, dispacth interface{}) *PpeConn {
	newPpeConn := &PpeConn{conn: conn,
		send:       make(chan bool, 100),
		shutDown:   make(chan bool),
		isShutDown: false}

	newPpeConn.register(dispacth)
	go newPpeConn.dispatchRoutine()
	go newPpeConn.sendRoutine()
	go newPpeConn.shutDownRoutine()
	return newPpeConn
}

func (self *PpeConn) shutDownRoutine() {
	<-self.shutDown
	self.conn.Close()
	fmt.Println("shut down!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func (self *PpeConn) SendUint16(data uint16) {
	if self.isShutDown {
		return
	}
	self.lsend.Lock()
	defer self.lsend.Unlock()
	binary.Write(&self.sendBuf, binary.LittleEndian, data)
	self.send <- true
}

func (self *PpeConn) Send(data []byte) {
	if self.isShutDown {
		return
	}
	self.lsend.Lock()
	defer self.lsend.Unlock()

	self.sendBuf.Write(data)
	self.send <- true
}

func (self *PpeConn) ShutDown() {
	if self.isShutDown {
		return
	}
	self.isShutDown = true
	self.shutDown <- true
}

func (self *PpeConn) register(arg interface{}) {

	//fmt.Println("开始注册")

	self.regitesr = arg
	self.methods = make(map[string]*methodType)
	self.msgs = make(map[uint16]*cmdType)

	// Install the methods
	rgType := reflect.TypeOf(arg)
	//fmt.Println("开始注册", rgType, rgType.NumMethod())
	msgId := uint16(1)
	for c := 0; c < rgType.NumMethod(); c++ {

		m := rgType.Method(c)
		mt := m.Type
		numP := mt.NumIn()

		if numP != 3 {
			continue
		}

		argType := mt.In(1)
		connType := mt.In(2)

		if connType.Elem().Name() != "PpeConn" {
			continue
		}

		fmt.Println("RPC注册成功:", msgId, m.Name, argType.Elem().Name())

		self.msgs[msgId] = &cmdType{rtType: argType.Elem()}
		self.methods[argType.Elem().Name()] = &methodType{method: m}
		msgId++
	}

	if len(self.msgs) == 0 {
		fmt.Println("没有任何合法有效的RPC回调函数成功注册")
	}
}

func (self *PpeConn) sendRoutine() {
	defer func() { self.ShutDown() }()
	for {
		<-self.send
		self.lsend.Lock()
		var tembuf bytes.Buffer
		tembuf.Write(self.sendBuf.Bytes())
		self.sendBuf.Reset()
		self.lsend.Unlock()

		_, err := tembuf.WriteTo(self.conn)

		if err != nil {
			break
		}
	}
}

func (self *PpeConn) dispatchRoutine() {
	defer func() {
		self.ShutDown()
		if r := recover(); r != nil {
			logger.Error("runtime error:", r)
			debug.PrintStack()
		}
	}()

	temp := make([]byte, 1024) //2kb的缓冲区
	for {
		nRead, err := self.conn.Read(temp)
		if err != nil {
			fmt.Println("conn.Read error :", err)
			break
		}

		self.recvBuf.Write(temp[:nRead])
		ok := true
		for {
			if self.recvBuf.Len() < 4 {
				break
			}

			srcData := self.recvBuf.Bytes()
			msgId := []uint8{srcData[0], srcData[1]}
			dataLen := []uint8{srcData[2], srcData[3]}

			msgIdLE := binary.LittleEndian.Uint16(msgId)
			dataLenLE := binary.LittleEndian.Uint16(dataLen)

			//fmt.Println("2", nRead, msgIdLE, dataLenLE)

			if self.recvBuf.Len() < int(dataLenLE+4) {
				break
			}

			//fmt.Println("3", msgIdLE, dataLenLE, self.msgs)

			cmdType, ok := self.msgs[msgIdLE]

			if !ok {
				break
			}

			self.recvBuf.Next(4)
			rtType := cmdType.rtType
			ptr := reflect.New(rtType)

			//fmt.Println("4", msgIdLE, reflect.TypeOf(rtType), ptr)

			fmt.Println("4", self.recvBuf.Bytes(), self.recvBuf.String())

			err := json.Unmarshal(self.recvBuf.Bytes(), ptr.Interface())
			if err != nil {
				fmt.Println("反序列化失败", ptr)
				ok = false
				break
			}
			self.methods[rtType.Name()].method.Func.Call([]reflect.Value{reflect.ValueOf(self.regitesr), ptr, reflect.ValueOf(self)})
			self.recvBuf.Next(int(dataLenLE))

			fmt.Println("call rpc function end !!!")
		}

		if !ok {
			break
		}

	}
}
