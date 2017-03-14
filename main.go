package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
    "path"
	// "io"
	"io/ioutil"
	"log"
	"net/http"
	//"sort"
	//"strconv"
    "strings"
	// "time"
	//"math"
	"github.com/gin-gonic/gin"
	// "github.com/gin-gonic/gin/binding"
	// "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
	// "text/template"
)

// Global
var config_path = flag.String("c", "config.yaml", "Path to a config file")

type Config_t struct {
    Port string        `yaml:"port"`
    Stats_path string  `yaml:"soket_dir"`
}

var Conf = Config_t {}

type Soket_t struct {
    Name string     `json:"name"`
    Proto string    `json:"proto"`
    Queue int       `json:"queue"`
    MaxQueue int    `json:"max_queue"`
    Shared int      `json:"shared"`
    CanOffload int  `json:"can_offload"`
}

type App_t struct {
    ID int `json:"id"`
    Modifier1 int `json:"modifier1"`
    Mountpoint string `json:"mountpoint"`
    StartupTime int `json:"startup_time"`
    Requests int `json:"requests"`
    Exceptions int `json:"exceptions"`
    Chdir string `json:"chdir"`
}

type Core_t struct {
    ID int                  `json:"id"`
    Requests int            `json:"requests"`
    StaticRequests int      `json:"static_requests"`
    RoutedRequests int      `json:"routed_requests"`
    OffloadedRequests int   `json:"offloaded_requests"`
    WriteErrors int         `json:"write_errors"`
    ReadErrors int          `json:"read_errors"`
    InRequest int           `json:"in_request"`
    Vars []interface{}      `json:"vars"`
    ReqInfo struct {}       `json:"req_info"`
}

type Worker_t struct {
    ID int              `json:"id"`
    Pid int             `json:"pid"`
    Accepting int       `json:"accepting"`
    Requests int        `json:"requests"`
    DeltaRequests int   `json:"delta_requests"`
    Exceptions int      `json:"exceptions"`
    HarakiriCount int   `json:"harakiri_count"`
    Signals int         `json:"signals"`
    SignalQueue int     `json:"signal_queue"`
    Status string       `json:"status"`
    Rss int             `json:"rss"`
    Vsz int             `json:"vsz"`
    RunningTime int     `json:"running_time"`
    LastSpawn int       `json:"last_spawn"`
    RespawnCount int    `json:"respawn_count"`
    Tx int              `json:"tx"`
    AvgRt int           `json:"avg_rt"`
    Apps []App_t        `json:"apps"`
    Cores []Core_t      `json:"cores"`
}

type Uwsgi_json_t struct {
    Version string          `json:"version"`
    ListenQueue int         `json:"listen_queue"`
    ListenQueueErrors int   `json:"listen_queue_errors"`
    SignalQueue int         `json:"signal_queue"`
    Load int                `json:"load"`
    Pid int                 `json:"pid"`
    UID int                 `json:"uid"`
    Gid int                 `json:"gid"`
    Cwd string              `json:"cwd"`
    Locks []map[string] int `json:"locks"`
    Sockets []Soket_t       `json:"sockets"`
    Workers []Worker_t      `json:"workers"`
}

func ProvideJsonText(path string) []byte {
    b, err := ioutil.ReadFile(path)

    if(err != nil) {
        log.Panic(err)
    }
    return b
}

/**
 * Sanitizze string
 * Remove " " --> "_"
 */
func SanitizeField (val string) string {
    ret := strings.Replace(val," ","_",-1)
    ret = strings.ToLower(val)
    return ret
}

/**
 * @brief Take uwsgi json struct and
 */
func uWSGI_DataFormat(data Uwsgi_json_t) string {
    var StrBuilder bytes.Buffer

    const uwsgi_prefix = "uwsgi_"

    StrBuilder.Write([]byte (fmt.Sprintf("%s%s %s\r\n",uwsgi_prefix, "general_version", data.Version)))
    StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_listen_queue", data.ListenQueue)))
    StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_listen_queue_errors", data.ListenQueueErrors)))
    StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_signal_queue", data.SignalQueue)))
    StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_load", data.Load)))

    for _, arr := range(data.Locks) {
        for key, val := range(arr) {
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s_%s %d\r\n", uwsgi_prefix, "locks", SanitizeField(key), val)))
        }
    }

    for _, Socket := range(data.Sockets) {
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_queue", Socket.Queue)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_max_queue", Socket.MaxQueue)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_shared", Socket.Shared)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_can_off_load", Socket.CanOffload)))
    }

    for _,Worker := range(data.Workers) {
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_accepting", Worker.ID, Worker.Accepting)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_requests", Worker.ID, Worker.Requests)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_delta_requests ", Worker.ID, Worker.DeltaRequests)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_exceptions ", Worker.ID, Worker.Exceptions)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_harakiri_count ", Worker.ID, Worker.HarakiriCount)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signals", Worker.ID, Worker.Signals)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signal_queue", Worker.ID, Worker.SignalQueue)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %s\r\n",uwsgi_prefix, "workers_status", Worker.ID, Worker.Status)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_rss", Worker.ID, Worker.Rss)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_vsz", Worker.ID, Worker.Vsz)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_running_time", Worker.ID, Worker.RunningTime)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_last_spawn", Worker.ID, Worker.LastSpawn)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_respawn_count", Worker.ID, Worker.RespawnCount)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_tx", Worker.ID, Worker.Tx)))
        StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_avg_rt", Worker.ID, Worker.AvgRt)))

        for _, App := range(Worker.Apps) {
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_modifier1", Worker.ID, App.ID, App.Modifier1)))
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_startup_time", Worker.ID, App.ID, App.StartupTime)))
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_requests", Worker.ID, App.ID, App.Requests)))
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_exceptions", Worker.ID, App.ID, App.Exceptions)))
        }

        for _, Core := range(Worker.Cores) {
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_requests", Worker.ID, Core.ID, Core.Requests)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_static_requests", Worker.ID, Core.ID, Core.StaticRequests)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_routed_requests", Worker.ID, Core.ID, Core.RoutedRequests)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_offloaded_requests", Worker.ID, Core.ID, Core.OffloadedRequests)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_write_errors", Worker.ID, Core.ID, Core.WriteErrors)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_read_errors", Worker.ID, Core.ID, Core.ReadErrors)))
            StrBuilder.Write([]byte( fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_in_request", Worker.ID, Core.ID, Core.InRequest)))
        }
    }

    fmt.Print("+--------------------------------------------------------------------------------+\r\n\r\n")
    fmt.Print(StrBuilder.String())

    fmt.Print("+--------------------------------------------------------------------------------+\r\n\r\n")
    return ""
}

/**
 * @brief make reading of all stast soket
 */
func Uwsgi_stats_Read () (string, error) {
    files, err := ioutil.ReadDir(Conf.Stats_path)
    if err != nil {
        log.Fatalf("Error on:%s\r\n%v:",Conf.Stats_path, err)
    }

    for _, file := range files {
        curret_uwsgi_data := new(Uwsgi_json_t)

        full_path := path.Join(Conf.Stats_path,file.Name())

        text := ProvideJsonText(full_path)
        json.Unmarshal(text, curret_uwsgi_data)
        uWSGI_DataFormat(*curret_uwsgi_data)
    }

    return "",nil
}

func parse_config () {
    data,err := ioutil.ReadFile(*config_path)

    if err != nil {
        log.Fatalf("Impossible read file:%s\r\n Error%v", config_path, err)
    }

    err = yaml.Unmarshal([]byte(data), &Conf)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
}

func main() {
    // Read cofig file
	flag.Parse()
    parse_config()

	router := gin.Default()

    Uwsgi_stats_Read()

    router.GET("/metrics", GET_Handling)
	// router.POST("/alert/:chatid", POST_Handling)
//    router.Run(fmt.Sprintf(":%s", Conf.Port))
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
