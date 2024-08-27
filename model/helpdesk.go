package model

type HelpdeskRekap struct {
	All  int `json:"all,omitempty"`
	ToDo int `json:"todo,omitempty"`
	Done int `json:"done,omitempty"`
}
