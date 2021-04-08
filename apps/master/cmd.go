package main

import (
	"github.com/bensema/redisocket"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/urfave/cli"
)

var (
	rsender *redisocket.Sender
)

// master server
func master(c *cli.Context) {

	mc := getMasterConfig(c)

	b, err := ioutil.ReadFile(mc.PublicKeyLocation)
	if err != nil {
		logger.Warn(err)
	}
	public_pem, rsaKeyErr := jwt.ParseRSAPublicKeyFromPEM(b)
	if rsaKeyErr != nil {
		logger.Warnf("Did not start \"%sdecode\" api", mc.ApiPrefix)
	}

	/*redis init*/
	rpool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", mc.RedisAddr)
		if err != nil {
			return nil, err
		}
		_, err = c.Do("SELECT", mc.RedisDb)
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, nil
	}, 10)
	rpool.MaxIdle = mc.RedisMaxIdle
	rpool.MaxActive = mc.RedisMaxConn

	/*Test redis connect*/
	err = RedisTestConn(rpool.Get())
	if err != nil {
		logger.Fatal(err)
	}
	rsender = redisocket.NewSender(rpool)

	engine := gin.Default()
	e := engine.Group("/")
	//e.POST("/push/socket/:app_key/:socket_id", PushToSocket)
	//e.POST("/push/user/:app_key/:user_id", PushToUser)
	e.POST("/push/:app_key/:channel/:event", PushMessage)
	e.POST("/push_batch/:app_key", PushBatchMessage)
	e.POST("/push/:app_key", PushMessageByPattern)
	e.POST("/reload/channel/user/:app_key/:user_id", ReloadUserChannels)
	e.POST("/add/channel/user/:app_key/:user_id", AddUserChannels)
	e.GET("/:app_key/channels", GetAllChannel)
	e.GET("/:app_key/channels/count", GetAllChannelCount)
	e.GET("/:app_key/online/bychannel/:channel", GetOnlineByChannel)
	e.GET("/:app_key/online/bychannel/:channel/count", GetOnlineCountByChannel)
	e.GET("/:app_key/online", GetOnline)
	e.GET("/:app_key/online/count", GetOnlineCount)
	//e.GET("/ping", Ping)

	if rsaKeyErr == nil {
		e.POST("/decode", func(c *gin.Context) {
			DecodeJWT(c, public_pem)
		})
	}
	serverError := make(chan error, 1)

	go func() {
		logger.Error(engine.Run(mc.ApiListen))
	}()

	// block and listen syscall
	shutdow_observer := make(chan os.Signal, 1)
	t := template.Must(template.New("gusher master start msg").Parse(masterMsgFormat))
	t.Execute(os.Stdout, mc)
	signal.Notify(shutdow_observer, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-shutdow_observer:
		logger.Info("Receive signal")
	case err := <-serverError:
		logger.Warn(err)
	}

}

func runtimeStats() (m *runtime.MemStats) {
	m = &runtime.MemStats{}

	//log.Println("# goroutines: ", runtime.NumGoroutine())
	runtime.ReadMemStats(m)
	//log.Println("Memory Acquired: ", m.Sys)
	//log.Println("Memory Used    : ", m.Alloc)
	return m
}
