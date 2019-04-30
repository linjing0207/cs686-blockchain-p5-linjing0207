package data

import "encoding/json"

type RegisterData struct {
	AssignedId int32 `json:"assignedId"`  //26
	PeerMapJson string `json:"peerMapJson"`
}

/**
Create NewRegisterData instance
 */
func NewRegisterData(id int32, peerMapJson string) RegisterData {
	registerData := RegisterData{id, peerMapJson}
	return registerData
}

/**
Encode RegisterData to json
 */
func (data *RegisterData) EncodeToJson() (string, error) {
	jsonRegisterData, err := json.Marshal(data)
	return string(jsonRegisterData), err
}