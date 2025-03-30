package mycontrol

import (
	"log"

	mc "github.com/odch/go-mycontrol"
)

func ActivateUser(apiToken string) (status, description string, err error) {
	c := mc.NewClient(apiToken)
	return activateUser(c, apiToken)
}

func activateUser(c mc.Client, apiToken string) (status, description string, err error) {
	status = "success"
	log.Println(apiToken)
	_, err = c.GetToken()
	if err != nil {
		status = "failure"
		return
	}
	return
}
