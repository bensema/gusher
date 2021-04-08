package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var env *string
var (
	version             string
	compileDate         string
	name                string
	listenChannelPrefix string
	cmdMaster           = cli.Command{
		Name:    "master",
		Usage:   "start gusher.master server",
		Action:  master,
		Aliases: []string{"ma"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "env-file,e",
				Usage: "import env file",
			},
			cli.BoolFlag{
				Name:  "debug,d",
				Usage: "open debug mode",
			},
		},
	}
	logger          *Logger
	masterMsgFormat = "\nmaster mode start at \"{{.GetStartTime}}\"\tserver ip:\"{{.ExternalIp}}\"\tversion:\"{{.Version}}\"\tcomplie at \"{{.CompileDate}}\"\n" +
		"api_listen:\"{{.ApiListen}}\"\tapi_preifx:\"{{.ApiPrefix}}\"\n" +
		"redis_addr:\"{{.RedisAddr}}\"\t" + "redis_dbno:\"{{.RedisDb}}\"\n" +
		"redis_max_idle:\"{{.RedisMaxIdle}}\"\n" +
		"redis_max_conn:\"{{.RedisMaxConn}}\"\n" +
		"log_formatter:\"{{.LogFormatter}}\"\n" +
		"public_key_location:\"{{.PublicKeyLocation}}\"\n\n"
)

func init() {
	listenChannelPrefix = name + "."
	/*logger init*/
	logger = GetLogger()
}

func getMasterConfig(c *cli.Context) (mc MasterConfig) {
	envInit(c)
	mc = MasterConfig{}
	mc.Name = os.Getenv("GUSHER_NAME")
	if mc.Name == "" {
		logger.Fatal("empty env GUSHER_NAME")
	}
	mc.PublicKeyLocation = os.Getenv("GUSHER_PUBLIC_PEM_FILE")
	if mc.PublicKeyLocation == "" {
		logger.Fatal("empty env GUSHER_PUBLIC_PEM_FILE")
	}
	mc.RedisAddr = os.Getenv("GUSHER_REDIS_ADDR")
	if mc.RedisAddr == "" {
		logger.Fatal("empty env GUSHER_REDIS_ADDR")
	}
	var err error
	mc.RedisDb, err = strconv.Atoi(os.Getenv("GUSHER_REDIS_DBNO"))
	if err != nil {
		mc.RedisDb = 0
	}
	mc.RedisMaxIdle, err = strconv.Atoi(os.Getenv("GUSHER_REDIS_MAX_IDLE"))
	if err != nil {
		mc.RedisMaxIdle = 10
	}
	mc.RedisMaxConn, err = strconv.Atoi(os.Getenv("GUSHER_REDIS_MAX_CONN"))
	if err != nil {
		mc.RedisMaxConn = 100
	}
	mc.ApiListen = os.Getenv("GUSHER_MASTER_API_LISTEN")
	if mc.ApiListen == "" {
		logger.Fatal("empty env GUSHER_MASTER_API_LISTEN")
	}
	mc.ApiPrefix = os.Getenv("GUSHER_MASTER_URI_PREFIX")
	if mc.ApiPrefix == "" {
		logger.Fatal("empty env GUSHER_MASTER_URI_PREFIX")
	}

	var f logrus.Formatter
	if strings.ToLower(mc.LogFormatter) == "json" || mc.LogFormatter == "" {
		f = &logrus.JSONFormatter{}
	} else {
		f = &logrus.TextFormatter{}
	}
	logger.SetFormatter(f)

	mc.StartTime = time.Now()
	mc.CompileDate = compileDate
	mc.Version = version
	mc.ExternalIp, err = GetExternalIP()
	if err != nil {
		logger.Fatal("cant get ip")
	}
	return
}

func envInit(c *cli.Context) {
	/*env init*/
	if c.String("env-file") != "" {
		envfile := c.String("env-file")
		//flag.Parse()
		err := godotenv.Load(envfile)
		if err != nil {
			logger.Fatal(err)
		}
	}

	if c.Bool("debug") {
		logger.Logger.Level = logrus.DebugLevel
	} else {
		logger.Logger.Level = logrus.InfoLevel
	}

}

func main() {
	cli.AppHelpTemplate += "\nWEBSITE:\n\t\thttps://github.com/syhlion/gusher.cluster\n\n"
	gusher := cli.NewApp()
	gusher.Name = name
	gusher.Author = "Scott (syhlion)"
	gusher.Usage = "very simple to use http request push message to websocket and very easy to scale"
	gusher.UsageText = "gusher.cluster master [-e envfile] [-d]"
	gusher.Version = version
	gusher.Compiled = time.Now()
	gusher.Commands = []cli.Command{
		cmdMaster,
	}
	gusher.Run(os.Args)

}
