package python

import (
	"bytes"
	"encoding/json"
	"github.com/wisecoach/ml_chain/proto"
	"io"
	"net/http"
	"strconv"
)

type httpPythonClient struct {
	TaskId     string
	ApiBaseUrl string
	httpClient *http.Client
}

func New(config *Config) Client {
	client := &httpPythonClient{
		TaskId:     config.TaskId,
		ApiBaseUrl: config.ApiBaseUrl,
		httpClient: &http.Client{},
	}
	return client
}

func (h *httpPythonClient) Init(genesis *proto.TaskGenesis) error {
	marshal, err := json.Marshal(genesis)
	if err != nil {
		return err
	}
	_, err = h.post("/init/"+h.TaskId, marshal)
	if err != nil {
		return err
	}
	return nil
}

func (h *httpPythonClient) Aggregate(request *AggregateRequest) (*AggregateResponse, error) {
	marshal, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	respBodyBytes, err := h.post("/aggregate", marshal)
	if err != nil {
		return nil, err
	}
	resp := &AggregateResponse{}
	err = json.Unmarshal(respBodyBytes, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *httpPythonClient) Validate(request *ValidateRequest) (*ValidateResponse, error) {
	marshal, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	respBodyBytes, err := h.post("/validate/"+h.TaskId, marshal)
	if err != nil {
		return nil, err
	}
	resp := &ValidateResponse{}
	err = json.Unmarshal(respBodyBytes, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *httpPythonClient) Train(request *TrainRequest) (*TrainResponse, error) {
	params := make(map[string]string)
	params["iteration"] = strconv.Itoa(request.Iteration)
	params["cindex"] = request.Cindex
	params["model_hash"] = request.GlobalModel.ModelHash
	respBodyBytes, err := h.get("/train", params)
	if err != nil {
		return nil, err
	}
	resp := &TrainResponse{}
	err = json.Unmarshal(respBodyBytes, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *httpPythonClient) get(uri string, params map[string]string) ([]byte, error) {
	url := h.ApiBaseUrl + uri
	isFirst := true
	for key, value := range params {
		if isFirst {
			url += "?" + key + "=" + value
			isFirst = false
		} else {
			url += "&" + key + "=" + value
		}
	}
	req, err := http.NewRequest("get", url, bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (h *httpPythonClient) post(uri string, requestBody []byte) ([]byte, error) {
	url := h.ApiBaseUrl + uri
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
