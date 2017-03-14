package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
    "path"
    "os"
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

type StatsSoketConf_t struct {
		Domain string `yaml:"domain"`
		Soket string `yaml:"soket"`
}

type Config_t struct {
	Port int `yaml:"port"`
	SoketDir string `yaml:"soket_dir"`
	StatsSokets []StatsSoketConf_t `yaml:"stats_sokets"`
}

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


/**
 * @brief Flag config
 */
var config_path = flag.String("c", "config.yaml", "Path to a config file")

/**
 * @brief Configuration struct
 */
var Conf Config_t

/**
 * @brief Provide sigle interface, for get json from uWSGI
 *
 * This function implement some differet way to get json from uWSGI.
 * uWSGI allow you to get text from Unix soket, HTTP soket
 */
func ProvideJsonText(path string) []byte {
    b, err := ioutil.ReadFile(path)

    if(err != nil) {
        log.Panic("[PANIC] Impossible read file:%v", err)
    }
    return b
}

/**
 * @brief Sanitizze string
 * 
 * Remove:
 * -    " " --> "_"
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

    /**
     * Try respect prometheus line guide for write expoter:
     * https://prometheus.io/docs/practices/naming/
     */
    const uwsgi_prefix = "uwsgi_"
    StrBuilder.WriteString("# HELP uWSGI Server\r\n")

    StrBuilder.WriteString(fmt.Sprintf("%s%s %s\r\n",uwsgi_prefix, "general_version", data.Version))
    StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_listen_queue", data.ListenQueue))
    StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_listen_queue_errors", data.ListenQueueErrors))
    StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_signal_queue", data.SignalQueue))
    StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "general_load", data.Load))

    StrBuilder.WriteString("\r\n# HELP uWSGI looks\r\n")
    for _, arr := range(data.Locks) {
        for key, val := range(arr) {

            StrBuilder.Write([]byte (fmt.Sprintf("%s%s_%s %d\r\n", uwsgi_prefix, "locks", SanitizeField(key), val)))
        }
    }

    for _, Socket := range(data.Sockets) {
        StrBuilder.WriteString("\r\n# HELP uWSGI Soket\r\n")

        StrBuilder.WriteString(fmt.Sprintf("Soket:%s\r\n", Socket.Name))

        StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_queue", Socket.Queue))
        StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_max_queue", Socket.MaxQueue))
        StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_shared", Socket.Shared))
        StrBuilder.WriteString(fmt.Sprintf("%s%s %d\r\n",uwsgi_prefix, "sockets_can_off_load", Socket.CanOffload))
    }

    for _,Worker := range(data.Workers) {
        StrBuilder.WriteString(fmt.Sprintf("\r\n# HELP uWSGI Worker %d\r\n",Worker.ID))

        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_accepting", Worker.ID, Worker.Accepting))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_requests", Worker.ID, Worker.Requests))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_delta_requests ", Worker.ID, Worker.DeltaRequests))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_exceptions ", Worker.ID, Worker.Exceptions))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_harakiri_count ", Worker.ID, Worker.HarakiriCount))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signals", Worker.ID, Worker.Signals))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signal_queue", Worker.ID, Worker.SignalQueue))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %s\r\n",uwsgi_prefix, "workers_status", Worker.ID, Worker.Status))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_rss", Worker.ID, Worker.Rss))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_vsz", Worker.ID, Worker.Vsz))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_running_time", Worker.ID, Worker.RunningTime))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_last_spawn", Worker.ID, Worker.LastSpawn))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_respawn_count", Worker.ID, Worker.RespawnCount))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_tx", Worker.ID, Worker.Tx))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_avg_rt", Worker.ID, Worker.AvgRt))

        for _, App := range(Worker.Apps) {
            StrBuilder.WriteString("\r\n# HELP uWSGI App\r\n")

            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_modifier1", Worker.ID, App.ID, App.Modifier1))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_startup_time", Worker.ID, App.ID, App.StartupTime))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_requests", Worker.ID, App.ID, App.Requests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_exceptions", Worker.ID, App.ID, App.Exceptions))
        }

        for _, Core := range(Worker.Cores) {
            StrBuilder.WriteString("\r\n# HELP uWSGI Core\r\n")

            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_requests", Worker.ID, Core.ID, Core.Requests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_static_requests", Worker.ID, Core.ID, Core.StaticRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_routed_requests", Worker.ID, Core.ID, Core.RoutedRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_offloaded_requests", Worker.ID, Core.ID, Core.OffloadedRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_write_errors", Worker.ID, Core.ID, Core.WriteErrors))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_read_errors", Worker.ID, Core.ID, Core.ReadErrors))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_in_request", Worker.ID, Core.ID, Core.InRequest))
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
func ReadStatsSoket_uWSGI () (string, error) {
    for _, SoketConf := range (Conf.StatsSokets) {
        Curret_uWSGI_Data := new(Uwsgi_json_t)

        FullPath := path.Join(Conf.SoketDir, SoketConf.Soket)

        // Check path exist
        _,err := os.Stat(FullPath)
        if err != nil {
            if os.IsNotExist(err) {
                /* Error is not fatal, soket could be removed, uWSGI restared...
                 * Then log it, and continue */
                log.Println("[ERROR] Could not open %s", FullPath)
            } else {
                log.Fatalln("[FATAL] %v", err)
            }
            continue
        }

        text := ProvideJsonText(FullPath)
        json.Unmarshal(text, Curret_uWSGI_Data)
        uWSGI_DataFormat(*Curret_uWSGI_Data)
    }
    return "",nil
}

/**
 * @brief Parse yaml config passed as flag parameter
 *
 */
func ParseConfig () {
    data, err := ioutil.ReadFile(*config_path)

    if err != nil {
        log.Fatalf("[FATAL] Impossible read file:%s\r\n Error%v", config_path, err)
    }

    err = yaml.Unmarshal([]byte(data), &Conf)
    if err != nil {
        log.Fatalf("[FATAL] %v", err)
    }
}

/**
 * @brief callback handler GET request
 */
func GET_Handling(c *gin.Context) {
	c.String(http.StatusOK, "Get")
    return
}

/**
 * @brief Validate 
 */
func ValidateConfig () {
    _,err := ioutil.ReadDir(Conf.SoketDir)

    // Calculate full path

    // Fist validate soket dir path
    if err != nil {
        log.Fatalf("[FATAL] Error on:%s\r\n%v:",Conf.SoketDir, err)
    }

    // Check path fist start polling
    for _, SoketPath := range(Conf.StatsSokets) {

        // Calculate full path
        FullPath := path.Join(Conf.SoketDir, SoketPath.Soket)

        // Check path exist
        _,err := os.Stat(FullPath)
        if err != nil {
           if os.IsNotExist(err) {
           /* Error is not fatal, soket could be removed, uWSGI restared...
            * Then log it, and continue */
            log.Println("[ERROR] Could not open %s", FullPath)
            } else {
                log.Fatalln("[FATAL] %v", err)
            }
        continue
        }
    }
}

func main() {
    // Init
	flag.Parse()
    ParseConfig()
    ValidateConfig()
    // End init

    // Is here for debug reason
    ReadStatsSoket_uWSGI()

    /* Enable here for reactivate GIN 
	 router := gin.Default()
     router.GET("/metrics", GET_Handling)
     router.Run(fmt.Sprintf(":%s", Conf.Port))
     */
}
