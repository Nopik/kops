/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"fmt"
	"testing"

	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
)

func Test_ValidateClusterPositive(t *testing.T) {
	nodeList, err := dummyClient("true", "true").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 1}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err != nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ValidateClusterMasterAndNodeNotReady(t *testing.T) {
	nodeList, err := dummyClient("false", "false").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 1}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err == nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ValidateClusterComponents(t *testing.T) {
	nodeList, err := dummyClient("true", "true").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var component = make([]string, 1)
	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 1, ComponentFailures: component}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err == nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ValidateClusterPods(t *testing.T) {
	nodeList, err := dummyClient("true", "true").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pod = make([]string, 1)
	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 1, PodFailures: pod}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err == nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ValidateClusterNodeNotReady(t *testing.T) {
	nodeList, err := dummyClient("true", "false").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 1}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err == nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ValidateClusterMastersNotEnough(t *testing.T) {
	nodeList, err := dummyClient("true", "true").Core().Nodes().List(v1.ListOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validationCluster := &ValidationCluster{NodeList: nodeList, NodesCount: 1, MastersCount: 3}
	validationCluster, err = validateTheNodes("foo", validationCluster)

	if err == nil {
		printDebug(validationCluster)
		t.Fatalf("unexpected error: %v", err)
	}
}

func printDebug(validationCluster *ValidationCluster) {
	fmt.Printf("cluster - masters ready: %v, nodes ready: %v\n", validationCluster.MastersReady, validationCluster.NodesReady)
	fmt.Printf("mastersNotReady %v\n", len(validationCluster.MastersNotReadyArray))
	fmt.Printf("mastersCount %v, mastersReady %v\n", validationCluster.MastersCount, len(validationCluster.MastersReadyArray))
	fmt.Printf("nodesNotReady %v\n", len(validationCluster.NodesNotReadyArray))
	fmt.Printf("nodesCount %v, nodesReady %v\n", validationCluster.NodesCount, len(validationCluster.NodesReadyArray))

}

const NODE_READY = "nodeReady"

func dummyClient(masterReady string, nodeReady string) *fake.Clientset {
	return fake.NewSimpleClientset(makeNodeList(
		[]map[string]string{
			{
				"name":               "master1",
				"kubernetes.io/role": "master",
				NODE_READY:           masterReady,
			},
			{
				"name":               "node1",
				"kubernetes.io/role": "node",
				NODE_READY:           nodeReady,
			},
		},
	))
}

func dummyNode(nodeMap map[string]string) v1.Node {

	nodeReady := v1.ConditionFalse
	if nodeMap[NODE_READY] == "true" {
		nodeReady = v1.ConditionTrue
	}
	expectedNode := v1.Node{
		ObjectMeta: v1.ObjectMeta{
			Name: nodeMap["name"],
			Labels: map[string]string{
				"kubernetes.io/role": nodeMap["kubernetes.io/role"],
			},
		},
		Spec: v1.NodeSpec{},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:    v1.NodeOutOfDisk,
					Status:  v1.ConditionTrue,
					Reason:  "KubeletOutOfDisk",
					Message: "out of disk space",
				},
				{
					Type:    v1.NodeMemoryPressure,
					Status:  v1.ConditionFalse,
					Reason:  "KubeletHasSufficientMemory",
					Message: "kubelet has sufficient memory available",
				},
				{
					Type:    v1.NodeDiskPressure,
					Status:  v1.ConditionFalse,
					Reason:  "KubeletHasSufficientDisk",
					Message: "kubelet has sufficient disk space available",
				},
				{
					Type:    v1.NodeReady,
					Status:  nodeReady,
					Reason:  "KubeletReady",
					Message: "kubelet is posting ready status",
				},
			},
			NodeInfo: v1.NodeSystemInfo{
				MachineID:     "123",
				SystemUUID:    "abc",
				BootID:        "1b3",
				KernelVersion: "3.16.0-0.bpo.4-amd64",
				OSImage:       "Debian GNU/Linux 7 (wheezy)",
				//OperatingSystem:         goruntime.GOOS,
				//Architecture:            goruntime.GOARCH,
				ContainerRuntimeVersion: "test://1.5.0",
				//KubeletVersion:          version.Get().String(),
				//KubeProxyVersion:        version.Get().String(),
			},
			Capacity: v1.ResourceList{
				v1.ResourceCPU:       *resource.NewMilliQuantity(2000, resource.DecimalSI),
				v1.ResourceMemory:    *resource.NewQuantity(20E9, resource.BinarySI),
				v1.ResourcePods:      *resource.NewQuantity(0, resource.DecimalSI),
				v1.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:       *resource.NewMilliQuantity(1800, resource.DecimalSI),
				v1.ResourceMemory:    *resource.NewQuantity(19900E6, resource.BinarySI),
				v1.ResourcePods:      *resource.NewQuantity(0, resource.DecimalSI),
				v1.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeLegacyHostIP, Address: "127.0.0.1"},
				{Type: v1.NodeInternalIP, Address: "127.0.0.1"},
				{Type: v1.NodeHostName, Address: nodeMap["name"]},
			},
			// images will be sorted from max to min in node status.
			Images: []v1.ContainerImage{
				{
					Names:     []string{"gcr.io/google_containers:v3", "gcr.io/google_containers:v4"},
					SizeBytes: 456,
				},
				{
					Names:     []string{"gcr.io/google_containers:v1", "gcr.io/google_containers:v2"},
					SizeBytes: 123,
				},
			},
		},
	}
	return expectedNode
}

// MakeNodeList constructs api.NodeList from list of node names and a NodeResource.
func makeNodeList(nodes []map[string]string) *v1.NodeList {
	var list v1.NodeList
	for _, node := range nodes {
		list.Items = append(list.Items, dummyNode(node))
	}
	return &list
}
