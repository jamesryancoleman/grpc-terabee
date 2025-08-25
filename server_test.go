package terabee

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jamesryancoleman/bos/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestConvertUrl(t *testing.T) {
	cases := []string{
		"terabee://192.168.13.120?serial=b827eb430fde",
		"terabee://b827eb430fde.local",
	}

	for i, xref := range cases {
		serial, url, err := ConvertXrefUrl(xref)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("%d: (%s) %#v\n", i+1, serial, url)
	}
}

func TestParseTerabee(t *testing.T) {
	cases := []string{
		"terabee://192.168.13.120?serial=b827eb430fde",
		"terabee://b827eb430fde.local",
	}
	for _, url := range cases {
		tc, err := ParseTerabee(url)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("%s\n", tc.String())
	}
}

func TestParseNGet(t *testing.T) {
	cases := []string{
		"terabee://192.168.13.120?serial=b827eb430fde",
		"terabee://b827eb430fde.local",
	}
	for i, xref := range cases {
		tc, err := ParseTerabee(xref)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("%d: fetching (%s) %#v\n", i+1, tc.Serial, tc.String())
		occupancy, err := tc.GetNetOccupancy()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("%d: occupancy = %d\n", i, occupancy)
	}
}

func TestStartServer(t *testing.T) {
	StartServer("0.0.0.0:50069")
}
func TestGet(t *testing.T) {
	s := StartServer("0.0.0.0:50069")

	// set up connection to server
	conn, err := grpc.NewClient(s.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect >> %s", err.Error())
	}
	defer conn.Close()
	c := common.NewGetSetRunClient(conn)

	// issue Get rpc
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := c.Get(ctx, &common.GetRequest{
		Header: &common.Header{Src: "test.local", Dst: s.Addr},
		Keys: []string{
			"terabee://192.168.13.120?serial=b827eb430fde",
			// "terabee://b827eb430fde.local",
		}})
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	for i, p := range r.GetPairs() {
		if p.GetError() > 0 {
			fmt.Printf("pair %d: error %d '%s'\n", i, p.GetError(), p.GetErrorMsg())
		} else {
			fmt.Printf("pair %d: %v\n", i, p)
		}
	}
}
