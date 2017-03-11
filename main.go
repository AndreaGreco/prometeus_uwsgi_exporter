package main

import (
	// "bytes"
	"encoding/json"
	"flag"
	"fmt"
	// "io"
	"io/ioutil"
	"log"
	"net/http"
	//"sort"
	//"strconv"
	//"strings"
	// "time"
	//"math"
	"github.com/gin-gonic/gin"
	// "github.com/gin-gonic/gin/binding"
	// "github.com/go-telegram-bot-api/telegram-bot-api"
	//"gopkg.in/yaml.v2"
	// "text/template"
)

// Global
var config_path = flag.String("c", "config.yaml", "Path to a config file")
var listen_addr = flag.String("l", ":9115", "Listen address")
// var uWSGI_stats_dir = flag.String("f", "/run/uwsgi/stats/")
var debug = flag.Bool("d",false,"Debug template")



// type  vars_t {

// }

// type req_info {

// }

type cores_t struct {
    id                 int
    requests           int
    static_requests    int
    routed_requests    int
    offloaded_requests int
    write_errors       int
    read_errors        int
    in_request         int
    vars_t             interface{}
    req_info           interface{}
}

type apps_t struct {
    id           int
	modifier1    int
	mountpoint   string
    startup_time int
	requests     int
	exceptions   int
	chdir        int
}

type Worker_t struct {
    id             int      `json:"id"`
    pid            int      `json:"pid"`
    accepting      int      `json:"accepting"`
	requests       int      `json:"requests"`
	delta_requests int      `json:"delta_requests"`
	exceptions     int      `json:"exceptions"`
	harakiri_count int      `json:"harakiri_count"`
	signals        int      `json:"signals"`
	signal_queue   int      `json:"signal_queue"`
	status         int      `json:"status"`
	rss            int      `json:"rss"`
	vsz            int      `json:"vsz"`
	running_time   int      `json:"running_time"`
	last_spawn     int      `json:"last_spawn"`
	respawn_count  int      `json:"respawn_count"`
	tx             int      `json:"tx"`
	avg_rt         int      `json:"avg_rt"`
    apps[]         apps_t   `json:"apps"`
    cores[]        cores_t  `json:"cores"`
}


func ProvideJsonText(path string) {
    b, err := ioutil.ReadFile(path)
    var 

    if(err != nil) {
        log.Panic(err)
    }
    json.Unmarshal(b,
}

/**
 * @brief Function that convert json text to Prometheus metrics
 * This function convert text from json to prometheus metrics
 */
func UwsgiMetrics(path string) string {
    log.Printf("Soket Connection WIP")
    return path
}

func Uwsgi_stats_Read () (string, error) {
    files, err := ioutil.ReadDir(".")
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
        fmt.Println(file.Name())
    }

    return "",nil
}

func main() {
	flag.Parse()

	router := gin.Default()

    Uwsgi_stats_Read()

    router.GET("/metrics", GET_Handling)
	// router.POST("/alert/:chatid", POST_Handling)
	router.Run(*listen_addr)
}

func GET_Handling(c *gin.Context) {
	c.String(http.StatusOK, "Get")
    return
}

// func POST_Handling(c *gin.Context) {
//
// 		c.String(http.StatusOK, "Post")
// 	return
// }
