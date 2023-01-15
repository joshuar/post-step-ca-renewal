/*
Copyright © 2023 Joshua Rich <joshua.rich@gmail.com>
*/
package actions

type Action struct {
	Name      string   `json:"name"`
	Cert      string   `json:"cert"`
	Key       string   `json:"key"`
	FullChain string   `json:"fullchain"`
	Command   []string `json:"command"`
}

type AllActions struct {
	ActionList []Action
}

func (a *AllActions) Count() int {
	return len(a.ActionList)
}

func (a *AllActions) GetActionByIndex(i int) *Action {
	return &a.ActionList[i]
}
