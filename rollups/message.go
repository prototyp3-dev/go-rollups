package rollups

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type FinishResponse struct {
	Type string          `json:"request_type"`
	Data json.RawMessage `json:"data"`
}

type InspectResponse struct {
	Payload string `json:"payload"`
}

type AdvanceResponse struct {
	Metadata Metadata `json:"metadata"`
	Payload  string   `json:"payload"`
}

type Metadata struct {
	MsgSender   string `json:"msg_sender"`
	EpochIndex  uint64 `json:"epoch_index"`
	InputIndex  uint64 `json:"input_index"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   uint64 `json:"timestamp"`
}

type Finish struct {
	Status string `json:"status"`
}

type Report struct {
	Payload string `json:"payload"`
}

type Notice struct {
	Payload string `json:"payload"`
}

type Voucher struct {
	Destination string `json:"destination"`
	Payload     string `json:"payload"`
}

type Exception struct {
	Payload string `json:"payload"`
}

type IndexResponse struct {
	Index uint64 `json:"index"`
}

var rollup_server string

func SetRollupServer(server_address string) {
	rollup_server = server_address
}

func GetRollupServer() string {
	return rollup_server
}

func SendPost(endpoint string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, rollup_server+"/"+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return &http.Response{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return http.DefaultClient.Do(req)
}

func SendFinish(finish *Finish) (*http.Response, error) {
	body, err := json.Marshal(finish)
	if err != nil {
		return &http.Response{}, err
	}

	return SendPost("finish", body)
}

func SendReport(report *Report) (*http.Response, error) {
	body, err := json.Marshal(report)
	if err != nil {
		return &http.Response{}, err
	}

	return SendPost("report", body)
}

func SendNotice(notice *Notice) (*http.Response, error) {
	body, err := json.Marshal(notice)
	if err != nil {
		return &http.Response{}, err
	}

	return SendPost("notice", body)
}

func SendVoucher(voucher *Voucher) (*http.Response, error) {
	body, err := json.Marshal(voucher)
	if err != nil {
		return &http.Response{}, err
	}

	return SendPost("voucher", body)
}

func SendException(exception *Exception) (*http.Response, error) {
	body, err := json.Marshal(exception)
	if err != nil {
		return &http.Response{}, err
	}

	return SendPost("exception", body)
}
