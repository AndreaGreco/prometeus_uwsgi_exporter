package  main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
    "strings"
    "net"
    "sync"
)

// JSON Definition

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

type Caches_t struct {
    Name string `json:"name"`
    Hash string `json:"hash"`
    Hashsize int `json:"hashsize"`
    Keysize int `json:"keysize"`
    MaxItems int `json:"max_items"`
    Blocks int `json:"blocks"`
    Blocksize int `json:"blocksize"`
    Items int `json:"items"`
    Hits int `json:"hits"`
    Miss int `json:"miss"`
    Full int `json:"full"`
    LastModifiedAt int `json:"last_modified_at"`
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
    Cache []Caches_t        `json:"caches", omitempty`
    Sockets []Soket_t       `json:"sockets"`
    Workers []Worker_t      `json:"workers"`
}

type WorkerStatus int
const (
    cheap WorkerStatus = iota
    pause
    sig
    busy
    idle
)



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
func uWSGI_DataFormat(data Uwsgi_json_t, domain string)string {
    /**
     * Try respect prometheus line guide for write expoter:
     * https://prometheus.io/docs/practices/naming/
     */
    var StrBuilder bytes.Buffer
    const uwsgi_prefix = "uwsgi_"

    StrBuilder.WriteString("# HELP uWSGI Server\r\n")

    StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %s\r\n",uwsgi_prefix, "general_version",domain, data.Version))
    StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "general_listen_queue",domain, data.ListenQueue))
    StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "general_listen_queue_errors",domain, data.ListenQueueErrors))
    StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "general_signal_queue",domain, data.SignalQueue))
    StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "general_load",domain, data.Load))

    for _, arr := range(data.Locks) {
        StrBuilder.WriteString("\r\n# HELP uWSGI looks\r\n")
        for key, val := range(arr) {
            StrBuilder.Write([]byte (fmt.Sprintf("%s%s_%s{domain=\"%s\"} %d\r\n", uwsgi_prefix, "locks",domain, SanitizeField(key), val)))
        }
    }

    for _, Socket := range(data.Sockets) {
        StrBuilder.WriteString("\r\n# HELP uWSGI Soket\r\n")

        StrBuilder.WriteString(fmt.Sprintf("Soket:%s\r\n", Socket.Name))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "sockets_queue",domain, Socket.Queue))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "sockets_max_queue",domain, Socket.MaxQueue))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "sockets_shared",domain, Socket.Shared))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\"} %d\r\n",uwsgi_prefix, "sockets_can_off_load",domain, Socket.CanOffload))
    }

    for _,Cache := range(data.Cache) {
        StrBuilder.WriteString(fmt.Sprintf("\r\n# HELP uWSGI Chache Name:%s, Hash:%s\r\n", Cache.Name,Cache.Hash))

        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_max_items",Cache.Name, Cache.Hash, domain, Cache.MaxItems))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_items",Cache.Name, Cache.Hash, domain, Cache.Items))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_hits",Cache.Name, Cache.Hash, domain, Cache.Hits))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_miss",Cache.Name, Cache.Hash, domain, Cache.Miss))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_full",Cache.Name, Cache.Hash, domain, Cache.Full))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\r\n", uwsgi_prefix, "chache_last_modified_at",Cache.Name, Cache.Hash, domain, Cache.LastModifiedAt))
    }

    for _,Worker := range(data.Workers) {
        StrBuilder.WriteString(fmt.Sprintf("\r\n###############                  HELP uWSGI Worker %d                   ###############\r\n",Worker.ID))

        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_accepting",domain, Worker.ID, Worker.Accepting))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_requests",domain, Worker.ID, Worker.Requests))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_delta_requests ",domain, Worker.ID, Worker.DeltaRequests))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_exceptions ",domain, Worker.ID, Worker.Exceptions))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_harakiri_count ",domain, Worker.ID, Worker.HarakiriCount))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signals",domain, Worker.ID, Worker.Signals))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_signal_queue",domain, Worker.ID, Worker.SignalQueue))
        StrBuilder.WriteString(fmt.Sprintf("# HELP uWSGI worker status: %d cheap, %d pause, %d sig, %d busy, %d idle\r\n", cheap, pause, sig, busy,idle))

        // Better return number
        switch(Worker.Status) {
        case "cheap":
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_status",domain, Worker.ID, cheap))
        case "pause":
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_status",domain, Worker.ID, pause))
        case "sig":
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_status",domain, Worker.ID, sig))
        case "busy":
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_status",domain, Worker.ID, busy))
        case "idle":
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_status",domain, Worker.ID, idle))
        }

        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_rss",domain, Worker.ID, Worker.Rss))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_vsz",domain, Worker.ID, Worker.Vsz))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_running_time",domain, Worker.ID, Worker.RunningTime))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_last_spawn",domain, Worker.ID, Worker.LastSpawn))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_respawn_count",domain, Worker.ID, Worker.RespawnCount))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_tx",domain, Worker.ID, Worker.Tx))
        StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\"} %d\r\n",uwsgi_prefix, "workers_avg_rt",domain, Worker.ID, Worker.AvgRt))

        for _, App := range(Worker.Apps) {
            StrBuilder.WriteString("\r\n# HELP uWSGI App\r\n")

            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_modifier1",domain, Worker.ID, App.ID, App.Modifier1))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_startup_time",domain, Worker.ID, App.ID, App.StartupTime))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_requests",domain, Worker.ID, App.ID, App.Requests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", app=\"%d\"} %d\r\n",uwsgi_prefix, "apps_exceptions",domain, Worker.ID, App.ID, App.Exceptions))
        }

        for _, Core := range(Worker.Cores) {
            StrBuilder.WriteString("\r\n# HELP uWSGI Core\r\n")

            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_requests",domain, Worker.ID, Core.ID, Core.Requests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_static_requests",domain, Worker.ID, Core.ID, Core.StaticRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_routed_requests",domain, Worker.ID, Core.ID, Core.RoutedRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_offloaded_requests",domain, Worker.ID, Core.ID, Core.OffloadedRequests))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_write_errors",domain, Worker.ID, Core.ID, Core.WriteErrors))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_read_errors",domain, Worker.ID, Core.ID, Core.ReadErrors))
            StrBuilder.WriteString(fmt.Sprintf("%s%s{domain=\"%s\", workerid=\"%d\", core=\"%d\"} %d\r\n",uwsgi_prefix, "cores_in_request",domain, Worker.ID, Core.ID, Core.InRequest))
        }
        StrBuilder.WriteString("#######################################################################################\r\n")
    }
    return StrBuilder.String()
}



/**
 * @brief Get json text from File DEBUG
 */
func ProvideJsonTextFile(path string) []byte {
    b, err := ioutil.ReadFile(path)

    if(err != nil) {
        log.Panic("[PANIC] Impossible read file:%v", err)
    }
    return b
}

/**
 * @brief Get json text from Unix Soket
 */
func ProvideJsonTextFromUnixSoket(FullPath string) ([]byte,error) {
    if CheckUnixSoket(FullPath) {
        log.Printf("[ERROR] Impossible open UnixSoket %s\r\n", FullPath)
        return nil,nil
    }

    c,err := net.Dial("unix", FullPath)
     if err != nil {
        return nil,err
    }

    buf := make([]byte, 1024*16)
    nbyte,err := c.Read(buf)
    c.Close()
    return buf[:nbyte],nil
}


/**
 * @brief make reading of all stast soket
 */
func ReadStatsSoket_uWSGI () string {
    var AllMetrics bytes.Buffer
    queue := make(chan string, 1)
    var wg sync.WaitGroup

    fmt.Printf("[DEBUG  ] Map len:%d\r\n", len(FileMap))

    wg.Add(len(FileMap))
    for Domain, FullPath := range FileMap {

        go func(FullPath string, CurretDomain string) {
            Curret_uWSGI_Data := new(Uwsgi_json_t)

            text,err := ProvideJsonTextFromUnixSoket(FullPath)
            if err != nil{
                log.Printf("[ERROR] Cannot read soket:%v", err)
                wg.Done()
                return
            }

            err = json.Unmarshal(text, Curret_uWSGI_Data)
            if err!= nil{
                log.Printf("[ERROR] Cannol Unmarshal json:%v", err)
                wg.Done()
                return
            }
            queue <- uWSGI_DataFormat(*Curret_uWSGI_Data, CurretDomain)
        } (FullPath, Domain)
    }

    // Recive data from channel
    go func() {
        for response := range queue {
            AllMetrics.WriteString(response)
            wg.Done()
        }
    }()
    wg.Wait()
    return AllMetrics.String()
}

