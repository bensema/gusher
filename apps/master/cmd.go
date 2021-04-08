package main

import (
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "net/http/pprof"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/syhlion/greq"
	"github.com/syhlion/httplog"
	"github.com/syhlion/requestwork.v2"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
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
	rsender := redisocket.NewSender(rpool)

	/*api start*/
	apiListener, err := net.Listen("tcp", mc.ApiListen)
	if err != nil {
		logger.Fatal(err)
	}
	r := mux.NewRouter()

	sub := r.PathPrefix(mc.ApiPrefix).Subrouter()
	sub.HandleFunc("/push/socket/{app_key}/{socket_id}", PushToSocket(rsender)).Methods("POST")
	sub.HandleFunc("/push/user/{app_key}/{user_id}", PushToUser(rsender)).Methods("POST")
	sub.HandleFunc("/push/{app_key}/{channel}/{event}", PushMessage(rsender)).Methods("POST")
	sub.HandleFunc("/push_batch/{app_key}", PushBatchMessage(rsender)).Methods("POST")
	sub.HandleFunc("/push/{app_key}", PushMessageByPattern(rsender)).Methods("POST")
	sub.HandleFunc("/reload/channel/user/{app_key}/{user_id}", ReloadUserChannels(rsender)).Methods("POST")
	sub.HandleFunc("/add/channel/user/{app_key}/{user_id}", AddUserChannels(rsender)).Methods("POST")
	sub.HandleFunc("/{app_key}/channels", GetAllChannel(rsender)).Methods("GET")
	sub.HandleFunc("/{app_key}/channels/count", GetAllChannelCount(rsender)).Methods("GET")
	sub.HandleFunc("/{app_key}/online/bychannel/{channel}", GetOnlineByChannel(rsender)).Methods("GET")
	sub.HandleFunc("/{app_key}/online/bychannel/{channel}/count", GetOnlineCountByChannel(rsender)).Methods("GET")
	sub.HandleFunc("/{app_key}/online", GetOnline(rsender)).Methods("GET")
	sub.HandleFunc("/{app_key}/online/count", GetOnlineCount(rsender)).Methods("GET")
	sub.HandleFunc("/ping", Ping()).Methods("GET")
	if rsaKeyErr == nil {
		sub.HandleFunc("/decode", DecodeJWT(public_pem)).Methods("POST")
	}
	n := negroni.New()
	n.Use(httplog.NewLogger(true))
	n.UseHandler(r)
	serverError := make(chan error, 1)
	server := http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      n,
	}
	go func() {
		err := server.Serve(apiListener)
		serverError <- err
	}()
	go func() {
		logger.Error(http.ListenAndServe(":7799", nil))
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
