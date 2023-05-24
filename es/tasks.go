package es

import "fmt"

type TasksResponse struct {
	Nodes map[string]TaskNode `json:"nodes"`
}

type TaskNode struct {
	Name             string            `json:"name"`
	TransportAddress string            `json:"transport_address"`
	Host             string            `json:"host"`
	IP               string            `json:"ip"`
	Roles            []string          `json:"roles"`
	Attributes       map[string]string `json:"attributes"`
	Tasks            map[string]Task   `json:"tasks"`
}

type Task struct {
	Node               string                 `json:"node"`
	ID                 int64                  `json:"id"`
	Type               string                 `json:"type"`
	Action             string                 `json:"action"`
	StartTimeInMillis  int64                  `json:"start_time_in_millis"`
	RunningTimeInNanos int64                  `json:"running_time_in_nanos"`
	Cancellable        bool                   `json:"cancellable"`
	Cancelled          bool                   `json:"cancelled"`
	ParentTaskID       string                 `json:"parent_task_id"`
	Headers            map[string]interface{} `json:"headers"`
}

func GetTasks(host string, port int) (TasksResponse, error) {
	url := fmt.Sprintf("http://%s:%d/_tasks", host, port)

	var response TasksResponse
	if err := getJSONResponse(url, &response); err != nil {
		return TasksResponse{}, err
	}

	return response, nil
}
