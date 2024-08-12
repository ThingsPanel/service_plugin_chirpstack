package main

import (
	"log"
	"plugin_chirpstack/cache"
	httpclient "plugin_chirpstack/http_client"
	httpservice "plugin_chirpstack/http_service"
	"plugin_chirpstack/mqtt"
	"plugin_chirpstack/services"
	"strings"

	"github.com/spf13/viper"
)

func main() {

	conf()
	LogInIt()
	log.Println("Starting the application...")

	//初始化redis
	cache.RedisInit()
	// 启动mqtt客户端
	mqtt.InitClient()
	// 启动http客户端
	httpclient.Init()
	// 启动服务
	//go services.Start()
	go services.StartHttp(services.NewChirpStack().Init())

	// 启动http服务
	httpservice.Init()
	select {}
}
func conf() {
	log.Println("加载配置文件...")
	// 设置环境变量前缀
	viper.SetEnvPrefix("plugin_ctwing")
	// 使 Viper 能够读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("加载配置文件完成...")
}
