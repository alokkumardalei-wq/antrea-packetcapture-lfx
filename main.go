package main

import (
	
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	nodeName = os.Getenv("NODE_NAME")
	captures = map[string]*exec.Cmd{}
)

func main() {
	if nodeName == "" {
		panic("NODE_NAME not set")
	}

	// In-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: handlePodUpdate,
		DeleteFunc: handlePodDelete,
	})

	stopCh := make(chan struct{})
	factory.Start(stopCh)

	select {} // run forever
}

func handlePodUpdate(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)

	if newPod.Spec.NodeName != nodeName {
		return
	}

	oldVal := oldPod.Annotations["tcpdump.antrea.io"]
	newVal := newPod.Annotations["tcpdump.antrea.io"]

	podName := newPod.Name

	// Annotation added
	if oldVal == "" && newVal != "" {
		startCapture(podName, newVal)
	}

	// Annotation removed
	if oldVal != "" && newVal == "" {
		stopCapture(podName)
	}
}

func handlePodDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	stopCapture(pod.Name)
}

func startCapture(podName, val string) {
	if _, exists := captures[podName]; exists {
		return
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println("Invalid annotation value:", val)
		return
	}

	file := fmt.Sprintf("/capture/capture-%s.pcap", podName)

	cmd := exec.Command(
		"tcpdump",
		"-C", "1",
		"-W", strconv.Itoa(n),
		"-w", file,
	)

	fmt.Println("Starting tcpdump for pod:", podName)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Failed to start tcpdump:", err)
		return
	}

	captures[podName] = cmd
}

func stopCapture(podName string) {
	cmd, exists := captures[podName]
	if !exists {
		return
	}

	fmt.Println("Stopping tcpdump for pod:", podName)
	cmd.Process.Kill()
	delete(captures, podName)

	files, _ := filepath.Glob("/capture/capture-" + podName + "*")
	for _, f := range files {
		if strings.Contains(f, podName) {
			os.Remove(f)
		}
	}
}
