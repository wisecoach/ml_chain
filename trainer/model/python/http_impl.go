package python

import (
	"bytes"
	"encoding/json"
	"net/http"
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

func (h *httpPythonClient) Aggregate(request *AggregateRequest) (*AggregateResponse, error) {
	marshal, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	respBodyBytes, err := h.post("/aggregate/"+h.TaskId, marshal)
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
	marshal, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	respBodyBytes, err := h.post("/train/"+h.TaskId, marshal)
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
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}
