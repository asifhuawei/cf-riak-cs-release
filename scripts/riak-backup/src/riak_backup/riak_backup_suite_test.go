package riak_backup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRiakBackup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RiakBackup Suite")
}
