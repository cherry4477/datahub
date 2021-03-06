package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/asiainfoLDP/datahub/ds"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	//"strconv"
	"strings"
	//"syscall"
)

const GstrDpPath string = "/var/lib/datahub"

type UserInfo struct {
	userName string
	password string
	b64      string
}

var (
	User          = UserInfo{}
	UnixSock      = "/var/run/datahub.sock"
	Logged        = false
	pidFile       = "/var/run/datahub.pid"
	CmdHttpServer = "localhost:35600"
)

type Command struct {
	Name      string
	SubCmd    []Command
	Handler   func(login bool, args []string) error
	Desc      string
	NeedLogin bool
}

type MsgResp struct {
	Msg string `json:"msg"`
}

const (
	ResultOK         = 0
	ErrorInvalidPara = iota + 4000
	ErrorNoRecord
	ErrorSqlExec
	ErrorInsertItem
	ErrorUnmarshal
	ErrorIOUtil
	ErrorMarshal
	ErrorServiceUnavailable
	ErrorFileNotExist
	ErrorTagAlreadyExist
	ErrorDatapoolNotExits
	ErrorRemoveAllJobs
	ErrorUnAuthorization
	ErrorOverLength
	ErrorOpenFile
	ErrorStatFile
	ErrorNoDatapoolDriver
	ErrorOtherError
	ErrorUnknowError
	ErrorItemNotExist
	ErrorPublishedItemEmpty
	ErrorPulledTagEmpty
	ErrorDpConnect
	ErrorOutMaxLength
	ErrorDatapoolAlreadyExits
	InternalError
)

const (
	ServerErrResultCodeOk   = 0
	ServerErrResultCode5009 = 5009
	ServerErrResultCode5012 = 5012
	ServerErrResultCode5023 = 5023
	ServerErrResultCode1009 = 1009
	ServerErrResultCode1400 = 1400
	ServerErrResultCode1008 = 1008
	ServerErrResultCode4010 = 4010
	ServerErrResultCode1011 = 1011
)

const (
	NoConsumingPlan    = 0
	ExitsConsumingPlan = 1
	RepositoryNotExist = 2
	DataitemNotExist   = 3
	TagNotExist        = 4
	RepoOrItemNotExist = 5
	TagExist           = 6
	RepoOrItemExist    = 7
)

var (
	ErrMsgArgument         string = "DataHub : Invalid argument."
	ValidateErrMsgArgument string = "DataHub : The parameter after rm is in wrong format."
	ErrLoginFailed         string = "Error : login failed."
)

var Cmd = []Command{
	{
		Name:    "dp",
		Handler: Dp,
		SubCmd: []Command{
			{
				Name:    "create",
				Handler: DpCreate,
			},
			{
				Name:    "rm",
				Handler: DpRm,
			},
		},
		Desc: "Datapool management",
	},
	{
		Name:    "ep",
		Handler: Ep,
		SubCmd: []Command{
			{
				Name:    "rm",
				Handler: EpRm,
			},
		},
		Desc: "Entrypoint management",
	},
	{
		Name:    "job",
		Handler: Job,
		SubCmd: []Command{
			{
				Name:    "rm",
				Handler: JobRm,
			},
		},
		Desc: "Job management",
	},
	{
		Name:      "login",
		Handler:   Login,
		Desc:      "Login to the server",
		NeedLogin: true,
	},
	{
		Name:      "logout",
		Handler:   Logout,
		Desc:      "Logout from the server",
		NeedLogin: true,
	},

	{
		Name:      "pub",
		Handler:   Pub,
		Desc:      "Publish a dataitem or tag",
		NeedLogin: true,
	},

	{
		Name:      "pull",
		Handler:   Pull,
		Desc:      "Pull the data subscribed",
		NeedLogin: true,
	},
	{
		Name:    "repo",
		Handler: Repo,
		SubCmd: []Command{
			{
				Name:    "rm",
				Handler: ItemOrTagRm,
			},
		},
		Desc:      "Repository management",
		NeedLogin: true,
	},
	{
		Name:      "subs",
		Handler:   Subs,
		Desc:      "Subscription of the dataitem",
		NeedLogin: true,
	},
	{
		Name:    "version",
		Handler: Version,
		Desc:    "Datahub version information",
	},
}

func login(interactive bool) {
	if Logged {
		if interactive {
			fmt.Println("You are already logged in as", User.userName)
		}
		return
	}

}

func commToDaemon(method, path string, jsonData []byte) (resp *http.Response, err error) {
	//fmt.Println(method, "/api"+path, string(jsonData))

	req, err := http.NewRequest(strings.ToUpper(method), "/api"+path, bytes.NewBuffer(jsonData))

	if len(User.userName) > 0 {
		req.SetBasicAuth(User.userName, User.password)
	}

	/* else {
		req.Header.Set("Authorization", "Basic "+os.Getenv("DAEMON_USER_AUTH_INFO"))
	}
	*/
	conn, err := net.Dial("tcp", CmdHttpServer)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Datahub daemon not running? Use 'datahub --daemon' to start daemon.")
		os.Exit(2)
	}
	//client := &http.Client{}
	client := httputil.NewClientConn(conn, nil)
	return client.Do(req)
	/*
		defer resp.Body.Close()
		response = *resp
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	*/
}

func printDash(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("-")
	}
	fmt.Println()
}

func showResponse(resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error :", resp.StatusCode, string(body))
		return
	}

	msg := MsgResp{}
	body, _ := ioutil.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &msg); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("DataHub : %v\n", msg.Msg)
	}
}

func showError(resp *http.Response) {

	if resp.StatusCode == http.StatusMovedPermanently {
		fmt.Println(ErrMsgArgument)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	result := ds.Result{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error : Unknown error. Http response code :", resp.StatusCode)
	} else {
		fmt.Printf("Error : %v\n", result.Msg)
	}

}

func StopP2P() error {

	/*data, err := ioutil.ReadFile(pidFile)

	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("datahub is not running.")
		}
	} else {
		if pid, err := strconv.Atoi(string(data)); err == nil {
			return syscall.Kill(pid, syscall.SIGQUIT)
		}
	}*/

	_, err := commToDaemon("get", "/stop", nil)
	return err
}

func ShowUsage() {
	fmt.Println("Usage:\tdatahub COMMAND [arg...]")
	fmt.Println("\tdatahub COMMAND [ --help ]")
	fmt.Println("\tdatahub help [COMMAND]\n")
	fmt.Println("A client for DataHub to publish and pull data\n")
	fmt.Println("Commands:")
	for _, v := range Cmd {
		fmt.Printf("    %-10s%s\n", v.Name, v.Desc)
	}
	fmt.Printf("\nrun '%s COMMAND --help' for details on a command.\n", os.Args[0])
}
