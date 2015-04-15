package riak_backup

import (
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
)

type InstanceMetadata struct {
	ServiceInstanceGuid string        "service_instance_guid"
	BoundApps           []AppMetadata "bound_apps"
}

type AppMetadata struct {
	Name string
	Guid string
}

func NewMetadataFromFilename(filename string) InstanceMetadata {
	yamlString, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
	}

	metadata := InstanceMetadata{}
	yaml.Unmarshal([]byte(yamlString), &metadata)
	return metadata
}
