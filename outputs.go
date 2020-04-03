package ipc

import "encoding/json"

type Mode struct {
	Width   int
	Height  int
	Refresh int
}

type Output struct {
	Name        string
	Make        string
	Model       string
	Serial      string
	Active      bool
	Scale       float32
	Modes       []*Mode
	CurrentMode *Mode `json:"current_mode"`
}

func (c *Connection) GetOutputs() ([]*Output, error) {
	j, err := c.send(GET_OUTPUTS, "")
	if err != nil {
		return nil, err
	}
	var outputs []*Output
	err = json.Unmarshal(j, &outputs)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}


