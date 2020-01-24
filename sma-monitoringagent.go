/**
 * @package   sma-monitoring-agent
 * @copyright sma-monitoring-agent contributors
 * @license   GNU Affero General Public License (https://www.gnu.org/licenses/agpl-3.0.de.html)
 *
 * @todo lots of documentation
 *
 *
 * Windows Monitoring Agent wiht REST-API
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/itcomusic/winsvc"
	"github.com/shirou/gopsutil/cpu"
	"gopkg.in/ini.v1"
)

var (
	Version   = "1.0.0"
	BuildTime = "2015-08-01 UTC"
	GitHash   = ""
)

type Check struct {
	Output   string
	ExitCode int
}

type Win32_LogicalDisk struct {
	Name      string
	FreeSpace string
	Size      string
}

type Win32_Process struct {
	Name        string
	Caption     string
	Commandline string
}

type Win32_Processor struct {
	LoadPercentage int
	Name           string
}
type Win32_Service struct {
	Caption string
	Name    string
	State   string
}
type Win32_OperatingSystem struct {
	TotalVisibleMemorySize int
	FreePhysicalMemory     int
	TotalVirtualMemorySize int
	FreeVirtualMemory      int
}
type AgentVersion struct {
	Version   string
	BuildTime string
	GitHash   string
}

type Win32_ComputerSystem struct {
	Model                     string
	Manufacturer              string
	Name                      string
	Domain                    string
	NumberOfProcessors        int
	NumberOfLogicalProcessors int
	TotalPhysicalMemory       int
}

type Inventory struct {
	Model                     string
	Manufacturer              string
	Name                      string
	Domain                    string
	NumberOfProcessors        int
	NumberOfLogicalProcessors int
	TotalPhysicalMemory       int
	IdentifyingNumber         string
}

type Win32_ComputerSystemProduct struct {
	IdentifyingNumber string
}

type CoreUsage struct {
	CPUType []cpu.InfoStat
	Usage   []float64
}

func isAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := LoadIni()
		if useSecret, _ := cfg.Section("server").Key("useSecret").Bool(); useSecret {
			secret := cfg.Section("server").Key("secret").String()
			if len(r.Header["Token"]) == 0 || secret != r.Header["Token"][0] {
				w.Write([]byte("Not Authorized"))
				return
			}
			endpoint(w, r)

		} else {
			endpoint(w, r)
		}
	})
}

func DiskUsage(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_LogicalDisk

	dl := r.URL.Query()["name"]

	qu := "WHERE MediaType ='12'"

	if len(dl) > 0 {

		qu = "WHERE Name LIKE'%" + dl[0] + "%'"
	}
	q := wmi.CreateQuery(&dst, qu)

	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func ProcessList(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Process

	dl := r.URL.Query()["name"]
	cl := r.URL.Query()["commandline"]
	re := regexp.MustCompile(`([^\\\"])\\([^\\\"])`)

	qu := ""

	if len(dl) > 0 {
		dl[0] = strings.ReplaceAll(dl[0], "\\\"", "\"")
		dl[0] = re.ReplaceAllString(dl[0], "$1\\\\$2")
		qu = "WHERE Name LIKE '%" + dl[0] + "%'"
	}
	if len(cl) > 0 {
		cl[0] = strings.ReplaceAll(cl[0], "\\\"", "\"")
		cl[0] = re.ReplaceAllString(cl[0], "$1\\\\$2")
		qu = "WHERE commandline LIKE '%" + cl[0] + "%'"

	}
	if len(dl) > 0 && len(cl) > 0 {
		dl[0] = strings.ReplaceAll(dl[0], "\\\"", "\"")
		dl[0] = re.ReplaceAllString(dl[0], "$1\\\\$2")

		cl[0] = strings.ReplaceAll(cl[0], "\\\"", "\"")
		cl[0] = re.ReplaceAllString(cl[0], "$1\\\\$2")
		qu = "WHERE Name LIKE '%" + dl[0] + "%' AND  commandline LIKE '%" + cl[0] + "%'"
	}

	q := wmi.CreateQuery(&dst, qu)

	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func WinService(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Service

	dl := r.URL.Query()["name"]
	qu := "WHERE StartMode LIKE '%Auto%'"

	if len(dl) > 0 {

		qu = "WHERE Name LIKE '%" + dl[0] + "%'"
	}
	q := wmi.CreateQuery(&dst, qu)

	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func MemoryUsage(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func InventoryService(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_ComputerSystem

	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}
	dst[0].Model = strings.Trim(dst[0].Model, " ")
	dst[0].Manufacturer = strings.Trim(dst[0].Manufacturer, " ")

	var dstp []Win32_ComputerSystemProduct
	qp := wmi.CreateQuery(&dstp, "")
	errp := wmi.Query(qp, &dstp)
	if errp != nil {
		log.Println(errp)
	}

	dstp[0].IdentifyingNumber = strings.Trim(dstp[0].IdentifyingNumber, " ")
	inventory := Inventory{dst[0].Model, dst[0].Manufacturer, dst[0].Name, dst[0].Domain, dst[0].NumberOfProcessors, dst[0].NumberOfLogicalProcessors, dst[0].TotalPhysicalMemory, dstp[0].IdentifyingNumber}

	var jsonData []byte

	jsonData, err = json.Marshal(inventory)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	dst = nil
	dstp = nil
}

func CPUUsage(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Processor
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Println(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func CPUUsageByCore(w http.ResponseWriter, r *http.Request) {

	cpuStat, _ := cpu.Info()
	percentage, _ := cpu.Percent(0, true)

	space := regexp.MustCompile(`\s+`)
	cpuStat[0].ModelName = space.ReplaceAllString(cpuStat[0].ModelName, " ")

	usage := &CoreUsage{
		CPUType: cpuStat,
		Usage:   percentage}

	jsonData, _ := json.Marshal(usage)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func ShowVersion(w http.ResponseWriter, r *http.Request) {
	agent := AgentVersion{Version: Version, BuildTime: BuildTime, GitHash: GitHash}
	jsonData, _ := json.Marshal(agent)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

type Application struct {
	srv *http.Server
}

func main() {
	winsvc.Run(func(ctx context.Context) {
		app := New()
		if err := app.Run(ctx); err != nil {
			log.Printf("[ERROR] rest terminated with error, %s", err)
			return
		}

		log.Printf("[WARN] rest terminated")
	})
	// service has been just stopped, but process of the go has not stopped yet
	// that is why recommendation is to not write any logic
}

func New() *Application {

	cfg := LoadIni()
	port := cfg.Section("server").Key("port").String()

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		agent := AgentVersion{Version: Version, BuildTime: BuildTime, GitHash: GitHash}
		jsonData := []byte(`{"Output":"SMA-MonitoringAgent ` + agent.Version + `","ExitCode":0}`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	mux.Handle("/api/diskusage", isAuthorized(DiskUsage))
	mux.Handle("/api/cpuusage", isAuthorized(CPUUsage))
	mux.Handle("/api/cpuusagebycore", isAuthorized(CPUUsageByCore))
	mux.Handle("/api/memoryusage", isAuthorized(MemoryUsage))
	mux.Handle("/api/processlist", isAuthorized(ProcessList))
	mux.Handle("/api/services", isAuthorized(WinService))
	mux.Handle("/api/systeminfo", isAuthorized(InventoryService))
	mux.Handle("/api/exec", isAuthorized(ExecuteScript))
	mux.Handle("/api/version", isAuthorized(ShowVersion))

	return &Application{srv: server}
}

func (a *Application) Run(ctx context.Context) error {
	log.Print("[INFO] started REST-API server version: " + Version)

	go func() {
		defer log.Print("[WARN] shutdown REST-API server")
		// shutdown on context cancellation
		<-ctx.Done()
		c, _ := context.WithTimeout(context.Background(), time.Second*5)
		a.srv.Shutdown(c)
	}()
	cfg := LoadIni()
	port := cfg.Section("server").Key("port").String()
	protocol := cfg.Section("server").Key("protocol").String()
	cert := cfg.Section("server").Key("certificate").String()
	privkey := cfg.Section("server").Key("privatekey").String()

	log.Println("[INFO] started http server on port: " + port)
	if protocol == "https" {
		return a.srv.ListenAndServeTLS(cert, privkey)
	} else {
		return a.srv.ListenAndServe()
	}
}
func ExecuteScript(w http.ResponseWriter, r *http.Request) {

	var waitStatus syscall.WaitStatus
	name := r.URL.Query().Get("name")
	cfg := LoadIni()
	yes := cfg.Section("commands").HasKey(name)

	if yes == true {

		params := cfg.Section("commands").Key(name).String()
		params = "/c " + params
		args := strings.Fields(params)
		cmd := exec.Command("cmd", args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
			} else {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
			}
		} else {
			waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		}

		check := Check{
			Output:   strings.TrimSpace(string(out)),
			ExitCode: waitStatus.ExitStatus(),
		}

		jsonData, _ := json.Marshal(check)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	} else {
		jsonData := []byte(`{"Output":"Command does not exist","ExitCode":3}`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func LoadIni() (cfg *ini.File) {

	path := os.Getenv("AGENT_INI_PATH")

	cfg, err := ini.Load(path + "agent.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	return cfg
}
