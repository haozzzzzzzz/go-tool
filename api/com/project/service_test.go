package project

import (
	"fmt"
	"testing"
)

func TestProject_Save(t *testing.T) {
	config := &ServiceConfigFormat{
		Name:       "test",
		ServiceDir: "/Users/hao/Documents/Projects/Github/go_lambda_learning/src/github.com/haozzzzzzzz/go-tool/api/common/proj",
	}

	project := &Service{
		Config: config,
	}

	err := project.Init()
	if nil != err {
		t.Error(err)
		return
	}
}

func TestProject_Load(t *testing.T) {
	project := &Service{}
	err := project.Load("/Users/hao/Documents/Projects/Github/go_lambda_learning/src/github.com/haozzzzzzzz/go-tool/api/common/proj")
	if nil != err {
		t.Error(err)
		return
	}

	fmt.Println(project.Config)
}
