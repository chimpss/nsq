package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
	"github.com/nsqio/nsq/internal/lg"
	"github.com/nsqio/nsq/internal/version"
	"github.com/nsqio/nsq/nsqlookupd"
)

func nsqlookupdFlagSet(opts *nsqlookupd.Options) *flag.FlagSet {
	flagSet := flag.NewFlagSet("nsqlookupd", flag.ExitOnError)

	flagSet.String("config", "", "path to config file， 默认为空")
	flagSet.Bool("version", false, "print version string，是否打印版本信息")

	logLevel := opts.LogLevel
	flagSet.Var(&logLevel, "log-level", "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.String("log-prefix", "[nsqlookupd] ", "log message prefix")
	flagSet.Bool("verbose", false, "[deprecated] has no effect, use --log-level")

	flagSet.String("tcp-address", opts.TCPAddress, "<addr>:<port> to listen on for TCP clients，默认 0.0.0.0:4160")
	flagSet.String("http-address", opts.HTTPAddress, "<addr>:<port> to listen on for HTTP clients，默认 0.0.0.0:4161")
	flagSet.String("broadcast-address", opts.BroadcastAddress, "address of this lookupd node, (默认 to the OS hostname)")

	flagSet.Duration("inactive-producer-timeout", opts.InactiveProducerTimeout, "duration of time a producer will remain in the active list since its last ping, 默认 300 * time.Second")
	flagSet.Duration("tombstone-lifetime", opts.TombstoneLifetime, "duration of time a producer will remain tombstoned if registration remains，默认 45 * time.Second，（tombstoned：已经不用的topic）")

	return flagSet
}

type program struct {
	// example：
	//var once sync.Once
	//onceBody := func() {
	//   fmt.Println("Only once")
	//}
	//done := make(chan bool)
	//for i := 0; i < 10; i++ {
	//   go func() {
	//       once.Do(onceBody)
	//       done <- true
	//   }()
	//}
	//for i := 0; i < 10; i++ {
	//   <-done
	//}
	once       sync.Once // sync包的once对象：只执行一次动作的对象。
	nsqlookupd *nsqlookupd.NSQLookupd
}

func main() {
	prg := &program{}
	// 启动 nsqlookupd 使用了svc库，可以在不同的系统中启动，守护进程
	// 只要实现svc.Service接口即可
	// start之前，先执行init
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		logFatal("%s", err)
	}
}

func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

func (p *program) Start() error {
	// 获取nsqlookupd默认配置
	opts := nsqlookupd.NewOptions()

	flagSet := nsqlookupdFlagSet(opts)
	flagSet.Parse(os.Args[1:])

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("nsqlookupd"))
		os.Exit(0)
	}

	// 从文件读取的配置，默认是没有的
	var cfg map[string]interface{}
	configFile := flagSet.Lookup("config").Value.String()
	if configFile != "" {
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			logFatal("failed to load config file %s - %s", configFile, err)
		}
	}

	// 使用了options包，这个包的作用：通过命令行标志，配置文件和默认结构值来解析配置值
	options.Resolve(opts, flagSet, cfg)
	nsqlookupd, err := nsqlookupd.New(opts)
	if err != nil {
		// 打印错误并退出程序
		logFatal("failed to instantiate nsqlookupd", err)
	}
	p.nsqlookupd = nsqlookupd

	go func() {
		err := p.nsqlookupd.Main()
		if err != nil {
			p.Stop()
			os.Exit(1)
		}
	}()

	return nil
}

// 退出守护进程，不过这个error返回写的有点疑问
func (p *program) Stop() error {
	p.once.Do(func() {
		p.nsqlookupd.Exit()
	})
	return nil
}

func logFatal(f string, args ...interface{}) {
	lg.LogFatal("[nsqlookupd] ", f, args...)
}
