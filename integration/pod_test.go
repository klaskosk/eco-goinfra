//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/pod"
	"github.com/stretchr/testify/assert"
)

func TestPodWatchForPodsInNamespaceRunning(t *testing.T) {
	t.Parallel()
	client := clients.New("")
	assert.NotNil(t, client)

	err := pod.WatchForPodsInNamespaceRunning(client, "hive", time.Hour)
	t.Logf("err: %v", err)
	assert.Nil(t, err)
}
