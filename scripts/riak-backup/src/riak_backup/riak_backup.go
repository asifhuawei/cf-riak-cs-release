package riak_backup

import (
	"encoding/json"
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
)

type Organization struct {
	Entity struct {
		Name string
	}
}

type Spaces struct {
	NextUrl   string `json:"next_url"`
	Resources []Space
}

type Space struct {
	Metadata struct {
		Guid string
	}
	Entity struct {
		Name             string
		OrganizationGuid string `json:"organization_guid"`
	}
}

type ServiceInstances struct {
	Services []ServiceInstance
}

type ServicePlan struct {
	Service struct {
		Label string
	}
}

type ServiceInstance struct {
	Guid        string
	Name        string
	ServicePlan ServicePlan `json:"service_plan"`
}

type Bindings struct {
	NextUrl   string `json:"next_url"`
	Resources []Binding
}

type Binding struct {
	Entity struct {
		App App
	}
}

type App struct {
	Metadata struct {
		Guid string
	}
	Entity struct {
		Name string
	}
}

func Backup(cf CfClientInterface, s3cmd S3CmdClientInterface, backup_dir string) {
	spaces := fetchSpaces(cf)

	for space_idx, space := range spaces {
		space_guid := space.Metadata.Guid
		space_name := space.Entity.Name
		organization := fetchOrganization(cf, space.Entity.OrganizationGuid)
		organization_name := organization.Entity.Name

		fmt.Printf("=== PROCESSING SPACE %d OF %d: SPACE NAME: %s, ORG NAME: %s\n", space_idx+1, len(spaces), space_name, organization_name)

		service_instances_json := cf.GetServiceInstancesForSpace(space_guid)
		service_instances := &ServiceInstances{}
		json.Unmarshal([]byte(service_instances_json), service_instances)

		for service_instance_idx, service_instance := range service_instances.Services {
			fmt.Printf("=== PROCESSING SERVICE INSTANCE %d OF %d IN SPACE %s: ", service_instance_idx+1, len(service_instances.Services), space_name)

			if service_instance.ServicePlan.Service.Label == "p-riakcs" {
				space_dir := spaceDirectory(backup_dir, organization_name, space_name)
				os.MkdirAll(space_dir, 0777)

				fmt.Println("\n")

				service_instance_guid := service_instance.Guid
				service_instance_name := service_instance.Name
				instance_dir := space_dir + "/service_instances/" + service_instance_name
				os.MkdirAll(instance_dir, 0777)
				writeMetadataFile(backup_dir, cf, organization_name, space_name, service_instance_name, service_instance_guid)

				data_dir := instance_dir + "/data"
				os.MkdirAll(data_dir, 0777)

				bucket_name := bucketNameFromServiceInstanceGuid(service_instance_guid)
				s3cmd.FetchBucket(bucket_name, data_dir)
			} else {
				fmt.Println("Not of type p-riakcs. Skipping.\n")
			}
		}
	}
}

func bucketNameFromServiceInstanceGuid(service_instance_guid string) string {
	return "service-instance-" + service_instance_guid
}

func fetchOrganization(cf CfClientInterface, organization_guid string) Organization {
	organization_json := cf.GetOrganization(organization_guid)
	organization := &Organization{}
	json.Unmarshal([]byte(organization_json), organization)
	return *organization
}

func fetchSpaces(cf CfClientInterface) []Space {
	spaces := []Space{}
	next_url := "/v2/spaces"
	for next_url != "" {
		spaces_json := cf.GetSpaces(next_url)
		page := &Spaces{}
		json.Unmarshal([]byte(spaces_json), page)

		spaces = append(spaces, page.Resources...)
		next_url = page.NextUrl
	}
	return spaces
}

func fetchBindings(cf CfClientInterface, service_instance_guid string) []Binding {
	next_url := "/v2/service_instances/" + service_instance_guid + "/service_bindings?inline-relations-depth=1"
	bindings := []Binding{}
	for next_url != "" {
		bindings_json := cf.GetBindings(next_url)
		page := &Bindings{}
		json.Unmarshal([]byte(bindings_json), page)

		bindings = append(bindings, page.Resources...)
		next_url = page.NextUrl
	}
	return bindings
}

func writeMetadataFile(backup_dir string, cf CfClientInterface, organization_name string, space_name string, service_instance_name string, service_instance_guid string) {
	bindings := fetchBindings(cf, service_instance_guid)

	metadata := InstanceMetadata{
		ServiceInstanceGuid: service_instance_guid,
	}

	app_metadatas := []AppMetadata{}
	for _, binding := range bindings {
		bound_app := binding.Entity.App
		app_metadatas = append(app_metadatas, AppMetadata{Name: bound_app.Entity.Name, Guid: bound_app.Metadata.Guid})
	}
	metadata.BoundApps = app_metadatas

	bytes, err := yaml.Marshal(metadata)
	if err != nil {
		fmt.Println(err.Error())
	}

	space_dir := spaceDirectory(backup_dir, organization_name, space_name)
	path := fmt.Sprintf("%s/service_instances/%s/metadata.yml", space_dir, service_instance_name)
	ioutil.WriteFile(path, bytes, 0777)
}

func spaceDirectory(backup_dir, organization_name, space_name string) string {
	return fmt.Sprintf("%s/orgs/%s/spaces/%s", backup_dir, organization_name, space_name)
}
