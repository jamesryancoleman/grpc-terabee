package terabee

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/jamesryancoleman/bos/common"
	"google.golang.org/grpc"
)

var serialRe = regexp.MustCompile(`^([a-zA-Z0-9]+).local`)
var ipRe = regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`)

var (
	Linfo  *log.Logger
	Lwarn  *log.Logger
	Ldebug *log.Logger
	Lslog  *slog.Logger
)

const netOccPath string = "/wizard/api/get_counts"
const queryStr string = "format=json"

type TerabeeConn struct {
	Serial string // also the host name
	IP     string
}

func (tc TerabeeConn) GetNetOccupancy() (int, error) {
	url := tc.String()
	username := "people_counting_admin"
	password := fmt.Sprintf("%s--admin", tc.Serial)

	// Create a custom Transport that skips TLS verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create an HTTP client with the custom Transport
	client := &http.Client{Transport: tr}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, nil
	}

	// Set the basic authentication header
	req.SetBasicAuth(username, password)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Process the response
	fmt.Println("Status Code:", resp.StatusCode)
	// You can read the response body here as well
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0, err
	}

	fmt.Println("Response Body:", string(body))

	var payload GetPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshall payload: %s", body)
		fmt.Println(errMsg)
		fmt.Println(err.Error())
		return 0, err
	}
	return payload.GetNetOcc(), nil
}

func (c *TerabeeConn) String() string {
	if c.IP != "" {
		return fmt.Sprintf("https://%s%s?%s", c.IP, netOccPath, queryStr)
	} else if c.Serial != "" {
		return fmt.Sprintf("https://%s.local%s?%s", c.Serial, netOccPath, queryStr)
	}
	panic("Terabee Connection empty")
}

func ParseTerabee(s string) (TerabeeConn, error) {
	var tc TerabeeConn
	urlParts, err := url.Parse(s)
	if err != nil {
		return tc, err
	}
	if urlParts.Scheme != "terabee" {
		return tc, fmt.Errorf("invalid scheme '%s'", urlParts.Scheme)
	}
	queryValues := urlParts.Query()
	V, ok := queryValues["serial"]
	if ok {
		tc.Serial = V[0]
	} else {
		match := serialRe.FindStringSubmatch(urlParts.Host)
		if len(match) > 0 {
			tc.Serial = match[1]
		}
	}
	match := ipRe.FindStringSubmatch(urlParts.Host)
	if len(match) > 0 {
		tc.IP = match[0]
	}
	if tc.Serial == "" {
		return tc, fmt.Errorf("no serial found in '%s'", s)
	}

	return tc, nil

}

func ParseSerial(s string) (string, error) {
	var serial string
	urlParts, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if urlParts.Scheme != "terabee" {
		return "", fmt.Errorf("invalid scheme '%s'", urlParts.Scheme)
	}
	queryValues := urlParts.Query()
	V, ok := queryValues["serial"]
	if ok {
		serial = V[0]
	} else {
		match := serialRe.FindStringSubmatch(urlParts.Host)
		if len(match) > 0 {
			serial = match[1]
		}
	}
	if serial == "" {
		return "", fmt.Errorf("no serial found in '%s'", s)
	}
	return serial, nil
}

func ConvertXrefUrl(s string) (string, string, error) {
	urlParts, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}
	if urlParts.Scheme != "terabee" {
		return "", "", fmt.Errorf("invalid scheme '%s'", urlParts.Scheme)
	}
	urlParts.Scheme = "https"
	urlParts.Path = netOccPath
	var serial string
	// url.ParseQuery(urlParts.ry)
	queryValues := urlParts.Query()
	V, ok := queryValues["serial"]
	if ok {
		serial = V[0]
	} else {
		match := serialRe.FindStringSubmatch(urlParts.Host)
		if len(match) > 0 {
			serial = match[1]
		}
	}
	if serial == "" {
		return "", "", fmt.Errorf("no serial found in '%s'", s)
	}
	// fmt.Printf("%#v\n", queryValues)
	urlParts.RawQuery = queryStr
	return serial, urlParts.String(), nil
}

func StartServer(listenAddr string) *Server {
	server := Server{}
	server.Start(listenAddr)

	// set up the custom loggers
	f, err := os.OpenFile("terabee-driver.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	jsonHandler := slog.NewJSONHandler(f, nil)

	Linfo = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	Lwarn = log.New(os.Stdout, "WARN:", log.Ltime|log.Lmicroseconds)
	Lslog = slog.New(jsonHandler)
	Ldebug = log.New(os.Stdout, "DEBUG:", log.Ltime|log.Lmicroseconds) // |log.Lshortfile

	common.Linfo.Println("driver server ready")
	return &server
}

type Server struct {
	common.UnimplementedGetSetRunServer
	Addr string
}

func (s *Server) Start(listenAddr string) {
	s.Addr = listenAddr
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	// create the grpc server
	server := grpc.NewServer()
	common.RegisterGetSetRunServer(server, s)

	// log successs
	common.StartLogging()
	common.Lslog.Info("server started", "listenAddr", s.Addr)
	common.Linfo.Printf("devCtrl: event=%s listenAddr=%s", "serverStarted", s.Addr) // to stdout

	// start the blocking gRPC server in a go routine
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func (s *Server) Get(ctx context.Context, req *common.GetRequest) (*common.GetResponse, error) {
	keys := req.GetKeys()
	pairs := make([]*common.GetPair, len(keys))

	// fetch values
	for i, k := range keys {
		tc, err := ParseTerabee(k)
		if err != nil {
			errMsg := err.Error()
			common.Lwarn.Println(errMsg)
			pairs[i] = &common.GetPair{
				Error:    common.GetError_GET_ERROR_UNSPECIFIED.Enum(),
				ErrorMsg: &errMsg,
			}
			continue
		}
		occupancy, err := tc.GetNetOccupancy()
		if err != nil {
			errMsg := err.Error()
			pairs[i] = &common.GetPair{
				Error:    common.GetError_GET_ERROR_UNSPECIFIED.Enum(),
				ErrorMsg: &errMsg,
			}
			continue
		}
		pairs[i] = &common.GetPair{Key: k, Value: fmt.Sprint(occupancy)}
	}
	return &common.GetResponse{Pairs: pairs}, nil
}
