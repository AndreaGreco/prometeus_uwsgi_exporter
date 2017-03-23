package main

import (
	"flag"
    "path"
    "os"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	// "github.com/gin-gonic/gin/binding"
	"gopkg.in/yaml.v2"
 //   "time"
    "fmt"
 //   "net"
)

type StatsSoketConf_t struct {
		Domain string `yaml:"domain"`
		Soket string `yaml:"soket"`
}

type Config_t struct {
	Port int `yaml:"port"`
	SoketDir string `yaml:"soket_dir"`
    PIDPath string `yaml:"pidfile"`
	StatsSokets []StatsSoketConf_t `yaml:"stats_sokets"`
}


/**
 * @brief map domain, full path
 */
var FileMap map[string] string

/**
  * @brief Flag config
  */
var config_path = flag.String("c", "config.yaml", "Path to a config file")
var noPID = flag.Bool("n",false,"Not deploy PID file")

/**
 * @brief Configuration struct
 */
var Conf Config_t

/**
 * @brief Parse yaml config passed as flag parameter
 * @return False if found error
 */
func ParseConfig () {

    data, err := ioutil.ReadFile(*config_path)

    if err != nil {
        log.Fatalf("[FATAL] Impossible read file:%s Error%v", *config_path, err)
    }

    err = yaml.Unmarshal([]byte(data), &Conf)
    if err != nil {
        log.Fatalf("[FATAL] %v", err)
    }
}

/**
 * @brief Check if unix soket exist, and if file is Unix Soket
 * 
 */
func CheckUnixSoket(FullPath string) bool {
    FoundError := false

    // Check path exist
    _,err := os.Stat(FullPath)
    if err != nil {
        if os.IsNotExist(err) {
            /* Error is not fatal, soket could be removed, uWSGI restared, Then log it, and continue */
            log.Printf("[ERROR] Could not open %s\r\nThis error is not critical will be SKIP", FullPath)
            FoundError = true
        } else {
            log.Fatalf("[FATAL] %v\r\n", err)
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
func GET_Handling(c *gin.Context) {
    str := ReadStatsSoket_uWSGI()
    if str == "" {
        c.String(http.StatusInternalServerError,"")
    }

	c.String(http.StatusOK, str)
    return
}

/**
 * @brief Validate config file and fist open of FD
 */
func ValidateConfig () {
    FoundError := false
    FileMap = make(map[string] string)
    log.Println("[INFO  ] Start check configuration file")

    _,err := ioutil.ReadDir(Conf.SoketDir)
    // Calculate full path
    // Fist validate soket dir path
    if err != nil {
        log.Fatalf("[FATAL] Error on:%s %v",Conf.SoketDir, err)
    }

    // Check path fist start polling
    for _, SoketPath := range(Conf.StatsSokets) {
        // Calculate full path
        FullPath := path.Join(Conf.SoketDir, SoketPath.Soket)

        if CheckUnixSoket(FullPath) {
            FoundError = true
        }
        FileMap[SoketPath.Domain] = FullPath
    }

    if !FoundError {
        log.Println("[INFO  ] Configuration correct, no error detect")
    } else {
        log.Println("[INFO  ] Error found check log")
    }
}

/**
 *
 * @brief Deploy pid file
 * Deploy pid file, Correct true, else false
 */
func DeployPID() bool {
    /**
     * TODO: Demonize
     * For do a good job this part must be demonizzed with double fork, write pid in /run/PIDNO
     * Find way to handle http with GIN or other lib to hangle reload, and restart signals
     */
    if *noPID {
        return true
    }

    // So ugly but it work
    PID := os.Getpid()

    pidFile,err := os.Open(Conf.PIDPath)
    if err != nil {
        log.Printf("Impossible open file, %s\r\n", err)
        pidFile.Close()
        return false
    }

    pidFile.WriteString(string(PID))
    pidFile.Sync()
    pidFile.Close()

    return true
}

func main() {
    // Init
	flag.Parse()
    ParseConfig()
    ValidateConfig()

    if !DeployPID() {
        return
    }

    log.Printf("[INFO  ] Bin port:%d", Conf.Port)
    gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
    router.GET("/metrics", GET_Handling)
    router.Run(fmt.Sprintf(":%d", Conf.Port))
}

