package utils

import (
	"github.com/casbin/casbin/v2"
	"log"
)

var Enforcer *casbin.Enforcer

func InitCasbin() {
	var err error
	Enforcer, err = casbin.NewEnforcer("casbin/model.conf", "casbin/policy.csv")
	if err != nil {
		log.Fatalf("Casbin enforcer init failed: %v", err)
	}
	if err = Enforcer.LoadPolicy(); err != nil {
		log.Fatalf("Casbin load policy failed: %v", err)
	}
}
