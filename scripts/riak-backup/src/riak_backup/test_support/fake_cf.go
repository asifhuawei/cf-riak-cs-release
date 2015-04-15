package test_support

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
)

type FakeCfClient struct {
}

func (cf *FakeCfClient) GetSpaces(next_url string) string {
	page := extractPageNumber("spaces", next_url)
	filename := fmt.Sprintf("successful_get_spaces_response_page_%d.json", page)
	return readFixtureFile(filename)
}

func (cf *FakeCfClient) GetOrganization(organization_guid string) string {
	var filename string
	switch organization_guid {
	case "organization-guid-0":
		filename = "successful_get_organization_response.json"
	default:
		panic("fixture file not found")
	}

	return readFixtureFile(filename)
}

func (cf *FakeCfClient) GetServiceInstancesForSpace(space_guid string) string {
	var filename string
	switch space_guid {
	case "space-guid-0":
		filename = "successful_get_instances_for_space_0_response.json"
	case "space-guid-1":
		return "{}"
	case "space-guid-2":
		filename = "successful_get_instances_for_space_2_response.json"
	case "space-guid-3":
		filename = "successful_get_instances_for_space_3_response.json"
	default:
		panic("fixture file not found")
	}

	return readFixtureFile(filename)
}

func (cf *FakeCfClient) GetBindings(next_url string) string {
	service_instance_guid := extractServiceInstanceGuid(next_url)

	resource := fmt.Sprintf("service_instances/%s/service_bindings", service_instance_guid)
	page := extractPageNumber(resource, next_url)
	path := fmt.Sprintf("successful_get_bindings_for_service_instance_0_response_page_%d.json", page)

	var filename string
	switch service_instance_guid {
	case "service-instance-guid-0":
		filename = path
	case "service-instance-guid-1", "service-instance-guid-2", "service-instance-guid-3":
		return "{}"
	default:
		panic(fmt.Sprintf("fixture file not found for %s", service_instance_guid))
	}

	return readFixtureFile(filename)
}

func (cf *FakeCfClient) Login(user, password string) {
}

func readFixtureFile(filename string) string {
	bytes, err := ioutil.ReadFile(getFixturePath(filename))
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(bytes)
}

func getFixturePath(filename string) string {
	_, current_filename, _, _ := runtime.Caller(0)
	current_dir := filepath.Dir(current_filename)
	return fmt.Sprintf("%s/cf_responses/%s", current_dir, filename)
}

func extractPageNumber(resource, next_url string) int {
	format := `page=(\d+)`
	r, _ := regexp.Compile(format)
	result := r.FindStringSubmatch(next_url)

	if 2 == len(result) {
		page_number, err := strconv.Atoi(result[1])
		if err != nil {
			fmt.Println(err.Error())
		}
		return page_number
	} else {
		return 1
	}
}

func extractServiceInstanceGuid(next_url string) string {
	format := "/v2/service_instances/(.*)/"
	r, _ := regexp.Compile(format)
	result := r.FindStringSubmatch(next_url)

	if 2 == len(result) {
		return result[1]
	} else {
		panic("could not extract instance guid from next_url: " + next_url)
	}
}
