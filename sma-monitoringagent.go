/**
 * @package   sma-monitoring-agent
 * @copyright sma-monitoring-agent contributors
 * @license   GNU Affero General Public License (https://www.gnu.org/licenses/agpl-3.0.de.html)
 * @authors   https://github.com/continentale/sma-monitoring-agent/graphs/contributors
 * @todo lots of documentation
 *
 *
 * Windows Monitoring Agent with REST-API
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

	endpointMemoryMap map[string]map[string]string
)

type Check struct {
	Output        string
	InMemoryValue string
	ExitCode      int
}

type Win32_LogicalDisk struct {
	Name      string
	FreeSpace string
	Size      string
}

type Win32_Process struct {
	Name    string
	Caption string
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

/*
 * Function to validate authorization ouf our REST-API.
 * uses the useSecret param in agent.ini
 */
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

/*
 * DiskUsage is used to dertermine the current disk usage via WMI.
 * Data source is Win32_LogicalDisk
 * default filter is fixed disk with MediaType 12
 */
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
		log.Fatal(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

/*
 * ProcessList is used to dertermine if a program is running
 * Data source is WMI Win32_Process
 * without filter all processes are shown
 * it's possible to filter via Name attribute, wildcards allowed
 */

func ProcessList(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Process

	dl := r.URL.Query()["name"]
	qu := ""

	if len(dl) > 0 {
		qu = "WHERE Name LIKE '%" + dl[0] + "%'"
	}
	q := wmi.CreateQuery(&dst, qu)

	err := wmi.Query(q, &dst)
	if err != nil {
		log.Fatal(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

/*
 * Function  WinService shows status of windows services.
 * data source is WMI Win32_Service
 * if no param is set all services with type Autostart will be validated
 */
func WinService(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Service

	dl := r.URL.Query()["name"]
	qu := "WHERE StartMode LIKE '%Auto%'"

	if len(dl) > 0 {
		qu = "WHERE Name='" + dl[0] + "'"
	}
	q := wmi.CreateQuery(&dst, qu)

	err := wmi.Query(q, &dst)
	if err != nil {
		log.Fatal(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

/*
 * Function  MemoryUsage shows memory usage statistics.
 * data source is WMI Win32_OperatingSystem
 */
func MemoryUsage(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	dst[0].Model = strings.Trim(dst[0].Model, " ")
	dst[0].Manufacturer = strings.Trim(dst[0].Manufacturer, " ")

	var dstp []Win32_ComputerSystemProduct
	qp := wmi.CreateQuery(&dstp, "")
	errp := wmi.Query(qp, &dstp)
	if errp != nil {
		log.Fatal(errp)
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

/*
 * Function  CPUUsage shows CPU usage statistics.
 * data source is WMI Win32_Processor
 */
func CPUUsage(w http.ResponseWriter, r *http.Request) {

	var dst []Win32_Processor
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Fatal(err)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(dst)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

/*
 * Function  CPUUsageByCore shows detailed CPU usage statistics.
 * currently beta, can fail with multi cpu systems and a lot of cores..
 */

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

/*
 * Function ShowVersion displays the agent version
 */

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
	endpointMemoryMap = make(map[string]map[string]string)

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

func parseCommandArgs(url, name string, arguments []string) []string {
	cfg := LoadIni()
	for i := range arguments {
		if arguments[i][0] == '$' {
			param := cfg.Section(name).Key(strings.Replace(arguments[i], "$", "", -1)).String()
			if param == "DATE" || param == "JSON" {
				// param is true or json get value from memory
				arguments[i] = endpointMemoryMap[url][strings.Replace(arguments[i], "$", "", -1)]
			} else if strings.Replace(arguments[i], "$", "", -1) == cfg.Section(name).Key(arguments[i][1:len(arguments[i])-1]).String() {
				// Variable Name is the type. Get single value from the json
				tmpMap := make(map[string]string)
				err := json.Unmarshal([]byte(endpointMemoryMap[url]["JSON"]), &tmpMap)
				if err != nil {
					log.Println(err)
				}
				arguments[i] = tmpMap[arguments[i][1:len(arguments[i])-1]]
			} else {
				// get value from the key
				arguments[i] = param
			}
		}
	}
	return arguments
}

func ExecuteScript(w http.ResponseWriter, r *http.Request) {
	var waitStatus syscall.WaitStatus

	name := r.URL.Query().Get("name")
	cfg := LoadIni()
	yes := cfg.Section("commands").HasKey(name)
	var check Check
	if yes {

		params := cfg.Section("commands").Key(name).String()
		params = "/c " + params
		args := strings.Fields(params)

		arguments := r.URL.Query()["args"]

		if r.URL.Query().Get("variables") == "true" {
			parseCommandArgs(r.URL.String(), name, arguments)
		}

		args = append(args, arguments...)

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

		outString := string(out)

		check.Output = strings.TrimSpace(outString[:strings.Index(outString, "{{")] + outString[strings.Index(outString, "}}")+2:])
		check.ExitCode = waitStatus.ExitStatus()
		if strings.Index(outString, "{{") != -1 && strings.Index(outString, "}}") != -1 {
			check.InMemoryValue = outString[strings.Index(outString, "{{")+1 : strings.LastIndex(outString, "}}")+1]
		}

		jsonData, _ := json.Marshal(check)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	} else {
		jsonData := []byte(`{"Output":"Command does not exist","ExitCode":3}`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}

	// set arguments only when URL-Parameter is given
	if r.URL.Query().Get("variables") == "true" {

		if _, isSet := endpointMemoryMap[r.URL.String()]["DATE"]; !isSet {
			endpointMemoryMap[r.URL.String()] = make(map[string]string)
		}

		endpointMemoryMap[r.URL.String()]["JSON"] = check.InMemoryValue
		endpointMemoryMap[r.URL.String()]["DATE"] = fmt.Sprint(time.Now().Unix())
	}

}

/*
 * Function LoadIni  is used to load the agent.ini file
 * os variable AGENT_INI_PATH can be used to load it from a custom location.
 */
func LoadIni() (cfg *ini.File) {

	path := os.Getenv("AGENT_INI_PATH")

	cfg, err := ini.Load(path + "agent.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	return cfg
}
