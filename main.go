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
	// "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
    "net"
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

/**
  * @brief Flag config
  */
var config_path = flag.String("c", "config.yaml", "Path to a config file")

/**
 * @brief Configuration struct
 */
var Conf Config_t

/**
 * @brief Map with active file descriptor
 */
var Active_FD map[string]net.Conn



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
 * @brief Open Unix soket and push it to UnixSoket fd map
 */
func OpenUnixFD(FullPath string, Domain string) error {
    c,err := net.Dial("unix",FullPath)
    if err != nil {
        log.Printf("[Error] Impossible open UnixSoket %s\r\n", FullPath)
        return err
    }
    Active_FD[Domain] = c
    log.Println("[INFO] Insert %s, domain:%s to polling list", FullPath, Domain)
    return nil
}

/**
 * @brief Close broken FD
 */
func CloseUnixFD(Domain string) {
    delete(Active_FD,Domain)
    log.Println("[INFO] Removed %s from polling list", Domain)
}

/**
 * @brief Validate config file and fist open of FD
 */
func ValidateConfig () {
    FoundError := false

    // Alloc File descriptor map
    Active_FD = make(map[string]net.Conn)

    log.Println("[INFO] Start check configuration file")

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
           /* Error is not fatal, soket could be removed, uWSGI restared, Then log it, and continue */
            log.Printf("[ERROR] Could not open %s\r\nThis error is not critical will be SKIP", FullPath)
            FoundError = true
            continue
            } else {
                log.Fatalf("[FATAL] %v\r\n", err)
                // Close here
            }

            // Let open File descriptor, check error

            err = OpenUnixFD(FullPath, SoketPath.Domain)

            err = OpenUnixFD(FullPath, SoketPath.Domain)
            if err != nil {
                FoundError = true
            }
        }
    }

    if !FoundError {
        log.Println("[INFO] Configuration correct, no error detect")
    }else {
        log.Println("[INFO] Error found check log")
    }
}

func main() {
    // Init
	flag.Parse()
    ParseConfig()
    ValidateConfig()
    // End init

    // Is here for debug reason
    ReadStatsSoket_uWSGI(&Active_FD)

    /* Enable here for reactivate GIN 
	 router := gin.Default()
     router.GET("/metrics", GET_Handling)
     router.Run(fmt.Sprintf(":%s", Conf.Port))
     */
}

