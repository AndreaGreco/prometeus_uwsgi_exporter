package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
)

// JSON Definition

type Socket_t struct {
	Name       string `json:"name"`
	Proto      string `json:"proto"`
	Queue      int    `json:"queue"`
	MaxQueue   int    `json:"max_queue"`
	Shared     int    `json:"shared"`
	CanOffload int    `json:"can_offload"`
}

type App_t struct {
	ID          int    `json:"id"`
	Modifier1   int    `json:"modifier1"`
	Mountpoint  string `json:"mountpoint"`
	StartupTime int    `json:"startup_time"`
	Requests    int    `json:"requests"`
	Exceptions  int    `json:"exceptions"`
	Chdir       string `json:"chdir"`
}

type Caches_t struct {
	Name           string `json:"name"`
	Hash           string `json:"hash"`
	Hashsize       int    `json:"hashsize"`
	Keysize        int    `json:"keysize"`
	MaxItems       int    `json:"max_items"`
	Blocks         int    `json:"blocks"`
	Blocksize      int    `json:"blocksize"`
	Items          int    `json:"items"`
	Hits           int    `json:"hits"`
	Miss           int    `json:"miss"`
	Full           int    `json:"full"`
	LastModifiedAt int    `json:"last_modified_at"`
}

type Core_t struct {
	ID                int           `json:"id"`
	Requests          int           `json:"requests"`
	StaticRequests    int           `json:"static_requests"`
	RoutedRequests    int           `json:"routed_requests"`
	OffloadedRequests int           `json:"offloaded_requests"`
	WriteErrors       int           `json:"write_errors"`
	ReadErrors        int           `json:"read_errors"`
	InRequest         int           `json:"in_request"`
	Vars              []interface{} `json:"vars"`
	ReqInfo           struct{}      `json:"req_info"`
}

type Worker_t struct {
	ID            int      `json:"id"`
	Pid           int      `json:"pid"`
	Accepting     int      `json:"accepting"`
	Requests      int      `json:"requests"`
	DeltaRequests int      `json:"delta_requests"`
	Exceptions    int      `json:"exceptions"`
	HarakiriCount int      `json:"harakiri_count"`
	Signals       int      `json:"signals"`
	SignalQueue   int      `json:"signal_queue"`
	Status        string   `json:"status"`
	Rss           int      `json:"rss"`
	Vsz           int      `json:"vsz"`
	RunningTime   int      `json:"running_time"`
	LastSpawn     int      `json:"last_spawn"`
	RespawnCount  int      `json:"respawn_count"`
	Tx            int      `json:"tx"`
	AvgRt         int      `json:"avg_rt"`
	Apps          []App_t  `json:"apps"`
	Cores         []Core_t `json:"cores"`
}

type Uwsgi_json_t struct {
	Version           string           `json:"version"`
	ListenQueue       int              `json:"listen_queue"`
	ListenQueueErrors int              `json:"listen_queue_errors"`
	SignalQueue       int              `json:"signal_queue"`
	Load              int              `json:"load"`
	Pid               int              `json:"pid"`
	UID               int              `json:"uid"`
	Gid               int              `json:"gid"`
	Cwd               string           `json:"cwd"`
	Locks             []map[string]int `json:"locks"`
	Cache             []Caches_t       `json:"caches", omitempty`
	Sockets           []Socket_t       `json:"sockets"`
	Workers           []Worker_t       `json:"workers"`
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
 * String buffer
 */
var StrBuilder bytes.Buffer

var globalDisableHelp bool

/**
 * @brief allow to enable \ref globalDisableHelp
 * Help must be redered only one time, or prometheus return parsing error
 * Then is disabled by render function every time is completed.busy
 * This function allow main thread to enable help at fist turn in for-loop
 */
func EnableHelp() {
	globalDisableHelp = true
}

/**
 * @brief uwsgi prefix string
 * string on head to all uwsgi metrics
 */
const uwsgi_prefix = "uwsgi_"

/**
 * @brief Sanitizze string
 *
 * Some metrics label must be sanitizzed, it contain space, or other special charter.
 * Remove:
 * -    " " --> "_"
 */
func SanitizeField(val string) string {
	ret := strings.Replace(val, " ", "_", -1)
	ret = strings.ToLower(ret)
	return ret
}

/**
 * @brief Write a help string in sting buffer
 */
func WriteHelp(str string, write bool) {
	if write && globalDisableHelp {
		StrBuilder.WriteString(str)
	}
}

/**
 * @brief Write metrics in string buffer
 */
func WriteMetrics(metrics string) {
	StrBuilder.WriteString(metrics)
}

/**
 * @brief Take uwsgi json struct and render Prometheus metrics
 */
func uWSGI_DataFormat(data Uwsgi_json_t, domain string) string {
	StrBuilder.Reset()

	/**
	 * Try respect prometheus line guide for write exporter:
	 * https://prometheus.io/docs/practices/naming/
	 */

	txt := "general_listen_queue"
	WriteHelp(fmt.Sprintf("# HELP %s%s Length of uwsgi listen queue\n", uwsgi_prefix, txt), true)
	WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, data.ListenQueue))

	txt = "general_listen_queue_errors"
	WriteHelp(fmt.Sprintf("# HELP %s%s Number of uwsgi server queue error\n", uwsgi_prefix, txt), true)
	WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, data.ListenQueueErrors))

	txt = "general_signal_queue"
	WriteHelp(fmt.Sprintf("# HELP %s%s Uwsgi signal queue\n", uwsgi_prefix, txt), true)
	WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, data.SignalQueue))

	for _, arr := range data.Locks {
		for key, val := range arr {
			key = SanitizeField(key)
			WriteHelp((fmt.Sprintf("# HELP %s%s_%s Locks %s\n", uwsgi_prefix, "locks", key, key)), true)
			WriteMetrics(fmt.Sprintf("%s%s_%s{domain=\"%s\"} %d\n", uwsgi_prefix, "locks", key, domain, val))
		}
	}

	for iSocket, Socket := range data.Sockets {
		txt := "sockets_queue"
		WriteHelp(fmt.Sprintf("# HELP %s%s Sockets queue\n", uwsgi_prefix, txt), (iSocket == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, Socket.Queue))

		txt = "sockets_max_queue"
		WriteHelp(fmt.Sprintf("# HELP %s%s Max queue length\n", uwsgi_prefix, txt), (iSocket == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, Socket.MaxQueue))

		txt = "sockets_shared"
		WriteHelp(fmt.Sprintf("# HELP %s%s Sockets shared\n", uwsgi_prefix, txt), (iSocket == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, Socket.Shared))

		txt = "sockets_can_off_load"
		WriteHelp(fmt.Sprintf("# HELP %s%s Sockets queue\n", uwsgi_prefix, txt), (iSocket == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\"} %d\n", uwsgi_prefix, txt, domain, Socket.CanOffload))
	}

	for iCache, Cache := range data.Cache {
		txt := "cache_max_items"
		WriteHelp(fmt.Sprintf("# HELP %s%s Max item in uwsgi cache\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.MaxItems))

		txt = "cache_items"
		WriteHelp(fmt.Sprintf("# HELP %s%s number current item in cache\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.Items))

		txt = "cache_hits"
		WriteHelp(fmt.Sprintf("# HELP %s%s cache hits\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.Hits))

		txt = "cache_miss"
		WriteHelp(fmt.Sprintf("# HELP %s%s Cache miss\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.Miss))

		txt = "cache_full"
		WriteHelp(fmt.Sprintf("# HELP %s%s Cache full\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.Full))

		txt = "cache_last_modified_at"
		WriteHelp(fmt.Sprintf("# HELP %s%s Last modified ithem in uwsgi cache\n", uwsgi_prefix, txt), (iCache == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", name=\"%s\", hash=\"%s\"} %d\n", uwsgi_prefix, txt, Cache.Name, Cache.Hash, domain, Cache.LastModifiedAt))
	}

	for iworker, Worker := range data.Workers {
		txt := "workers_id"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker id\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.ID))

		txt = "workers_accepting"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker accepting request\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Accepting))

		txt = "workers_requests"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker request elaborated\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Requests))

		txt = "workers_delta_requests"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker delta request TODO: Check uwsgiSource\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.DeltaRequests))

		txt = "workers_exceptions"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker exceptions counter\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Exceptions))

		txt = "workers_harakiri_count"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker harakiri count\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.HarakiriCount))

		txt = "workers_signals"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker signal elaborated count\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Signals))

		txt = "workers_signal_queue"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker signal queue\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.SignalQueue))

		WriteHelp(fmt.Sprintf("# HELP %s%s Worker status %d cheap %d pause %d sig %d busy %d idle\n", uwsgi_prefix, "workers_status", cheap, pause, sig, busy, idle), (iworker == 0))
		// Better return number
		switch Worker.Status {
		case "cheap":
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, "workers_status", domain, iworker, cheap))
		case "pause":
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, "workers_status", domain, iworker, pause))
		case "sig":
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, "workers_status", domain, iworker, sig))
		case "busy":
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, "workers_status", domain, iworker, busy))
		case "idle":
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, "workers_status", domain, iworker, idle))
		}

		txt = "workers_rss"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker resident set size\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Rss))

		txt = "workers_vsz"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker virtual memory sizen", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Vsz))

		txt = "workers_running_time"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker Worker running time\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.RunningTime))

		txt = "workers_last_spawn"
		WriteHelp(fmt.Sprintf("# HELP %s%s Last time stamp spawn\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.LastSpawn))

		txt = "workers_respawn_count"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker respawn count\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.RespawnCount))

		txt = "workers_tx"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker byte trasmitted\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.Tx))

		txt = "workers_avg_rt"
		WriteHelp(fmt.Sprintf("# HELP %s%s Worker average request\n", uwsgi_prefix, txt), (iworker == 0))
		WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, Worker.AvgRt))

		for _, App := range Worker.Apps {
			txt := "apps_modifier1"
			WriteHelp(fmt.Sprintf("# HELP %s%s Apps modifier1 wsgi protocol\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", app=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, App.ID, App.Modifier1))

			txt = "apps_startup_time"
			WriteHelp(fmt.Sprintf("# HELP %s%s Apps start up time\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", app=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, App.ID, App.StartupTime))

			txt = "apps_requests"
			WriteHelp(fmt.Sprintf("# HELP %s%s App request\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", app=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, App.ID, App.Requests))

			txt = "apps_exceptions"
			WriteHelp(fmt.Sprintf("# HELP %s%s App execeptions\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", app=\"%d\"} %d\n", uwsgi_prefix, txt, domain, iworker, App.ID, App.Exceptions))
		}

		for _, Core := range Worker.Cores {
			txt := "cores_requests"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores requests\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_requests", domain, iworker, Core.ID, Core.Requests))

			txt = "cores_static_requests"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores static resource request count\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_static_requests", domain, iworker, Core.ID, Core.StaticRequests))

			txt = "cores_routed_requests"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores routed requests\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_routed_requests", domain, iworker, Core.ID, Core.RoutedRequests))

			txt = "cores_offloaded_requests"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores offloaded requests\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_offloaded_requests", domain, iworker, Core.ID, Core.OffloadedRequests))

			txt = "cores_write_errors"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores write errors\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_write_errors", domain, iworker, Core.ID, Core.WriteErrors))

			txt = "cores_read_errors"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores read errors\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_read_errors", domain, iworker, Core.ID, Core.ReadErrors))

			txt = "cores_in_request"
			WriteHelp(fmt.Sprintf("# HELP %s%s Cores in request\n", uwsgi_prefix, txt), (iworker == 0))
			WriteMetrics(fmt.Sprintf("%s%s{domain=\"%s\", workerenum=\"%d\", core=\"%d\"} %d\n", uwsgi_prefix, "cores_in_request", domain, iworker, Core.ID, Core.InRequest))
		}
	}
	// Disable help must be write only fist time or prometeus return error
	globalDisableHelp = false
	return StrBuilder.String()
}

/**
 * @brief Get json text from File DEBUG
 */
func ProvideJsonTextFile(path string) []byte {
	b, err := ioutil.ReadFile(path)

	if err != nil {
		log.Criticalf("Impossible read file:%v\n", err)
	}
	return b
}

/**
 * @brief Get json text from Unix Socket
 */
func ProvideJsonTextFromUnixSocket(FullPath string) ([]byte, error) {
	if CheckUnixSocket(FullPath) {
		log.Errorf("Impossible open UnixSocket %s\r\n", FullPath)
		return nil, nil
	}

	c, err := net.Dial("unix", FullPath)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024*16)
	nbyte, err := c.Read(buf)
	c.Close()
	return buf[:nbyte], nil
}

type uWSGI_BufferCollector struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

/**
 * @brief Perform reading of uwsgi stast socket
 */
func ReadStatsSocket_uWSGI() []byte {
	/*
	 * Worning, this function use go subroutine for parallel reading of socket.
	 * Function is splitted this part:
	 * -   Create comunication channel
	 * -   Create all go-routing with for loop
	 * -   Read all unix socket
	 * -   Concatenate string
	 * -   close channel
	 * This part is sheduled by internal go scheduler.
	 */
	var wg sync.WaitGroup

	var AllMetrics uWSGI_BufferCollector
	AllMetrics.buffer.Reset()

	log.Debugf("Map len:%d\n", len(FileMap))

	// Enable help fist time
	EnableHelp()
	wg.Add(len(FileMap))
	for Domain, FullPath := range FileMap {
		go func(FullPath string, CurretDomain string) {
			Curret_uWSGI_Data := new(Uwsgi_json_t)

			text, err := ProvideJsonTextFromUnixSocket(FullPath)
			if err != nil {
				log.Errorf("Cannot read socket:%v", err)
				wg.Done()
				return
			}

			err = json.Unmarshal(text, Curret_uWSGI_Data)
			if err != nil {
				log.Errorf("Cannon Unmarshal json:%v", err)
				wg.Done()
				return
			}

			// Formatting routine is not thread safe group all in global, then look here
			AllMetrics.mutex.Lock()
			value := uWSGI_DataFormat(*Curret_uWSGI_Data, CurretDomain)
			AllMetrics.buffer.WriteString(value)
			AllMetrics.mutex.Unlock()

			wg.Done()
		}(FullPath, Domain)
	}
	wg.Wait()
	return AllMetrics.buffer.Bytes()
}
