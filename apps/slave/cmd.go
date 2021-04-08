package main

import (
	"github.com/bensema/redisocket"
	"html/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/urfave/cli"
)

//slave server
func slave(c *cli.Context) {

	sc := getSlaveConfig(c)
	/*redis init*/
	rpool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", sc.RedisAddr)
		if err != nil {
			return nil, err
		}
		_, err = c.Do("SELECT", sc.RedisDb)
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, nil
	}, 10)

	rpool.MaxIdle = sc.RedisMaxIdle
	rpool.MaxActive = sc.RedisMaxConn
	rpool.Wait = true
	rpool.IdleTimeout = 240 * time.Second
	rpool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		if time.Since(t) < time.Minute {
			return nil
		}
		_, err := c.Do("PING")
		return err
	}

	jobRpool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", sc.RedisJobAddr)
		if err != nil {
			return nil, err
		}
		_, err = c.Do("SELECT", sc.RedisJobDb)
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, nil
	}, 10)

	jobRpool.MaxIdle = sc.RedisJobMaxIdle
	jobRpool.MaxActive = sc.RedisJobMaxConn
	jobRpool.Wait = true
	jobRpool.IdleTimeout = 240 * time.Second
	jobRpool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		if time.Since(t) < time.Minute {
			return nil
		}
		_, err := c.Do("PING")
		return err
	}

	/*Test redis connect*/
	err := RedisTestConn(rpool.Get())
	if err != nil {
		logger.Fatal(err)
	}
	err = RedisTestConn(jobRpool.Get())
	if err != nil {
		logger.Fatal(err)
	}

	rsHub := redisocket.NewHub(rpool, logger.GetLogger(), c.Bool("debug"))
	rsHub.Config.MaxMessageSize = int64(sc.MaxMessage)
	rsHub.Config.ScanInterval = sc.ScanInterval
	rsHub.Config.Upgrader.ReadBufferSize = sc.ReadBuffer
	rsHub.Config.Upgrader.WriteBufferSize = sc.WriteBuffer
	rsHubErr := make(chan error, 1)
	go func() {
		rsHubErr <- rsHub.Listen(listenChannelPrefix)
	}()

	engine := gin.Default()
	e := engine.Group("/")

	e.GET("/ws/:app_key", func(c *gin.Context) {
		WsConnect(c, rsHub)
	})

	go func() {
		logger.Error(engine.Run(sc.ApiListen))
	}()

	defer func() {
		rsHub.Close()
		rpool.Close()
	}()

	// block and listen syscall
	shutdow_observer := make(chan os.Signal, 1)
	t := template.Must(template.New("gusher slave start msg").Parse(slaveMsgFormat))
	t.Execute(os.Stdout, sc)
	signal.Notify(shutdow_observer, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-shutdow_observer:
		logger.Info("receive signal")
	case err := <-rsHubErr:
		logger.Error("redis sub connection diconnect ", err)
	}
	return

}
