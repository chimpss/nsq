package nsqlookupd

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/nsqio/nsq/internal/http_api"
	"github.com/nsqio/nsq/internal/protocol"
	"github.com/nsqio/nsq/internal/util"
	"github.com/nsqio/nsq/internal/version"
)

type NSQLookupd struct {
	sync.RWMutex
	opts         *Options              // 配置
	tcpListener  net.Listener          // tcp监听地址
	httpListener net.Listener          // http监听地址
	waitGroup    util.WaitGroupWrapper // waitGroup方式的并发控制
	DB           *RegistrationDB       // 注册数据存储
}

func New(opts *Options) (*NSQLookupd, error) {
	var err error

	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, opts.LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	//
	l := &NSQLookupd{
		opts: opts,
		DB:   NewRegistrationDB(),
	}

	l.logf(LOG_INFO, version.String("nsqlookupd"))

	// 实例化tcpListener
	l.tcpListener, err = net.Listen("tcp", opts.TCPAddress)
	if err != nil {
		return nil, fmt.Errorf("listen (%s) failed - %s", opts.TCPAddress, err)
	}
	// 实例化httpListener
	l.httpListener, err = net.Listen("tcp", opts.HTTPAddress)
	if err != nil {
		return nil, fmt.Errorf("listen (%s) failed - %s", opts.TCPAddress, err)
	}

	return l, nil
}

// 入口函数
// Main starts an instance of nsqlookupd and returns an
// error if there was a problem starting up.
func (l *NSQLookupd) Main() error {
	ctx := &Context{l}

	exitCh := make(chan error)
	var once sync.Once
	exitFunc := func(err error) {
		once.Do(func() {
			if err != nil {
				l.logf(LOG_FATAL, "%s", err)
			}
			exitCh <- err
		})
	}

	tcpServer := &tcpServer{ctx: ctx}
	// 异步执行，tcp监听
	l.waitGroup.Wrap(func() {
		// 如果TCPServer函数错误，退出的话，则通过exitCh告知外面
		exitFunc(
			// tcp loop
			protocol.TCPServer(l.tcpListener, tcpServer, l.logf),
		)
	})
	httpServer := newHTTPServer(ctx)
	// 异步执行，http监听
	l.waitGroup.Wrap(func() {
		exitFunc(
			http_api.Serve(l.httpListener, httpServer, "HTTP", l.logf),
		)
	})

	// 等待两个异步结果，如果收到信号，则表示出问题了，返回err
	err := <-exitCh
	return err
}

func (l *NSQLookupd) RealTCPAddr() *net.TCPAddr {
	return l.tcpListener.Addr().(*net.TCPAddr)
}

func (l *NSQLookupd) RealHTTPAddr() *net.TCPAddr {
	return l.httpListener.Addr().(*net.TCPAddr)
}

// 关闭监听
// 执行条件：如果有一个listener异常了，那么久会执行此函数
// 而执行此函数，会关闭两个listener，如果关闭，就会使两个net loop都会异常
// 也就是说，只要有一个关闭了，就会导致两个都关闭。
// 这就解释了 l.waitGroup.Wait() 这行代码的作用
func (l *NSQLookupd) Exit() {
	if l.tcpListener != nil {
		l.tcpListener.Close()
	}

	if l.httpListener != nil {
		l.httpListener.Close()
	}

	// 等待两个tcp都结束之后，exit才成功
	l.waitGroup.Wait()
}
