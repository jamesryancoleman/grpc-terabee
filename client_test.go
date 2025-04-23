package terabee

import (
	"fmt"
	"testing"

	"github.com/jamesryancoleman/bos/common"
)

func TestGetOccupancy(t *testing.T) {
	serial := "b827eb430fde"
	occupancy, err := GetNetOccupancy(serial)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
	fmt.Printf("current occupancy is: %d\n", occupancy)
}

func TestCommonAccess(t *testing.T) {
	key := common.RedisKey{}
	fmt.Printf("%v!\n", key)
}
