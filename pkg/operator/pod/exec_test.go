package pod

import (
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

// @TODO: Think on how to test it
func TestExecCmdExample(t *testing.T) {
	t.Skip()
	clientSet := operator.BuildExternalClient()
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatalf("unable to get cluster config from flags, error %v", err)
	}

	namespace := "swarm"
	name := "swarm-worker-0"
	cmd := "cat /app/config/config.yml"
	r, err := ExecCmd(clientSet, config, namespace, name, cmd)
	if err != nil {
		t.Fatalf("unable to execute command, error %v", err)
	}

	//spew.Dump(outBuf.String(), errBuf.String())

	if err := os.WriteFile("foo.yaml", r, 0644); err != nil {
		t.Fatalf("unable to write fil, error %v", err)
	}

}
