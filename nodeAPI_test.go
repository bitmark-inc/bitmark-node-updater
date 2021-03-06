package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

///////////
// Stand alone test by using go test  nodeAPI_test.go nodeAPI.go types.go
// Run and Observe bitmark-node GUI
//////////

func TestBitmarkdStart(t *testing.T) {
	time.Sleep(6 * time.Second)
	body, err := NodeAPI("", bitmarkdStart)
	var resp BitmarkdServResp
	json.Unmarshal(body, &resp)
	fmt.Println(resp.Message)
	fmt.Println(resp.OK)
	assert.NoError(t, err, "NodeAPI TestBitmarkdStart error")

}

func TestBitmarkdStop(t *testing.T) {
	time.Sleep(6 * time.Second)
	body, err := NodeAPI("", bitmarkdStop)
	var resp BitmarkdServResp
	json.Unmarshal(body, &resp)
	fmt.Println(resp.Message)
	fmt.Println(resp.OK)
	assert.NoError(t, err, "NodeAPI TestBitmarkdStop error")

}
func TestRecorderdStart(t *testing.T) {
	time.Sleep(6 * time.Second)
	body, err := NodeAPI("", recorderdStart)
	var resp RecorderdServResp
	json.Unmarshal(body, &resp)
	fmt.Println(resp.Message)
	fmt.Println(resp.OK)
	assert.NoError(t, err, "NodeAPI TestRecorderdStart error")
}

func TestRecorderdStop(t *testing.T) {
	time.Sleep(6 * time.Second)
	body, err := NodeAPI("", recorderdStop)
	var resp RecorderdServResp
	json.Unmarshal(body, &resp)
	fmt.Println(resp.Message)
	fmt.Println(resp.OK)
	assert.NoError(t, err, "NodeAPI TestRecorderdStop error")
}

/* initialize Log for sigle file unit test
func initLog() {
	logfile, err := os.OpenFile("unittest.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	log.Init("", true, false, logfile)
}
*/
