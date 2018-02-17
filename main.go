package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
)

type StatsSocketConf_t struct {
	Domain string `yaml:"domain"`
	Socket string `yaml:"socket"`
}

type Config_t struct {
	Port         int                 `yaml:"port"`
	SocketDir    string              `yaml:"socket_dir"`
	PIDPath      string              `yaml:"pidfile"`
	StatsSockets []StatsSocketConf_t `yaml:"stats_sockets"`
}

// LOGGER
var log = logging.MustGetLogger("uwsg_exporter")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{shortfunc}  %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func SetUpLogger() {
	BkStdOut := logging.NewLogBackend(os.Stderr, "", 0)
	BkStdoutFormatter := logging.NewBackendFormatter(BkStdOut, format)
	// Set the backends to be used.
	logging.SetBackend(BkStdoutFormatter)
}

/**
 * @brief map domain, full path
 */
var FileMap map[string]string

/**
 * @brief Flag config
 */
var config_path = flag.String("c", "config.yaml", "Path to a config file")

/**
 * Do not deploy PID FileMap
 */
var noPID = flag.Bool("n", false, "Not deploy PID file")

/**
 * @brief Configuration struct
 */
var Conf Config_t

/**
 * @brief Parse yaml config passed as flag parameter
 * @return False if found error
 */
func ParseConfig() {
	data, err := ioutil.ReadFile(*config_path)

	if err != nil {
		log.Fatalf("Impossible read file:%s Error%v\n", *config_path, err)
	}

	err = yaml.Unmarshal([]byte(data), &Conf)
	if err != nil {
		log.Fatalf("Config file:%s Error, %v\n", *config_path, err)
	}
}

/**
 * @brief Check if unix socket exist, and if file is Unix Socket
 *
 */
func CheckUnixSocket(FullPath string) bool {
	FoundError := false

	// Check path exist
	_, err := os.Stat(FullPath)
	if err != nil {
		if os.IsNotExist(err) {
			/* Error is not fatal, socket could be removed, uWSGI restared, Then log it, and continue */
			log.Errorf("Could not open %s. This error is not critical will be SKIP\n", FullPath)
			FoundError = true
		} else {
			log.Fatalf("%v\n", err)
		}
	}
	// Let open File descriptor, check error
	if err != nil {
		FoundError = true
	}
	return FoundError
}

/**
 * @brief callback handler GET request
 */
func GET_Handling(w http.ResponseWriter, r *http.Request) {
	w.Write(ReadStatsSocket_uWSGI())
	w.Write([]byte(fmt.Sprintf("uwsgiexpoter_subroutine %d\n", runtime.NumGoroutine())))
}

/**
 * @brief Validate config file and fist open of FD
 */
func ValidateConfig() {
	FoundError := false
	FileMap = make(map[string]string)
	log.Info("Start check configuration file\n")

	// Check if folder exist
	_, err := ioutil.ReadDir(Conf.SocketDir)
	if err != nil {
		log.Fatalf("Error %v\n", err)
	}

	// Check path fist start polling
	for _, SocketPath := range Conf.StatsSockets {
		// Calculate full path
		var FullPath string

		if path.IsAbs(SocketPath.Socket) {
			// Support socket with absolute path
			FullPath = SocketPath.Socket
		} else {
			FullPath = path.Join(Conf.SocketDir, SocketPath.Socket)
		}

		log.Infof("Socket Path:%s", FullPath)

		if CheckUnixSocket(FullPath) {
			FoundError = true
		}

		FileMap[SocketPath.Domain] = FullPath
	}

	if !FoundError {
		log.Info("Configuration correct, no error detect\n")
	} else {
		log.Info("Error found check log\n")
	}
}

/**
 *
 * @brief Deploy pid file
 * Deploy pid file, Correct true, else false
 */
func DeployPID() bool {
	if *noPID {
		return true
	}

	// So ugly but it work
	PID := os.Getpid()

	pidFile, err := os.Open(Conf.PIDPath)
	if err != nil {
		log.Fatalf("Impossible open file, %s\n", err)
	}

	pidFile.WriteString(string(PID))
	pidFile.Sync()
	pidFile.Close()

	return true
}

func main() {
	// Setup env
	flag.Parse()
	ParseConfig()
	SetUpLogger()
	DeployPID()
	// End setup, from here all will be moved in second fork

	ValidateConfig()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", Conf.Port))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Bind port:%d\n", Conf.Port)
	http.HandleFunc("/metrics", GET_Handling)

	http.Serve(l, nil)
}
