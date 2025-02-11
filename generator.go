package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (

	// TODO: Remove it!

	configMapName = "podtinytidyid-counter"
	defaultValue  = "0"
)

// getNext gets the next Identifier for a podSet.
func GetNext(podSet string, bitsInID int) string {
	// Get the configmap
	var configmap *corev1.ConfigMap
	var nextId string
	var err error
	isIdValid := false

	klog.Infof("GetNext for: %s", podSet)
	for !isIdValid {
		configmap, err = getConfigMap()
		if err != nil {
			klog.Errorf("getting configmap: %s", err)
			continue
		}
		// Use the configmap variable here
		nextId, err = getNextInConfigmap(configmap, podSet, bitsInID)
		if err != nil {
			klog.Errorf("getting next: %s", err)
			continue
		}
		// Check if there is a pod with hte nextId as Id.
		// Get the pod with labelId
		err = podDoesNotExist(labelSet, podSet, getLabelId(podSet), nextId)
		if err != nil {
			klog.Errorf("getting pod: %s", err)
			continue
		}
		isIdValid = true
	}
	return nextId
}

func getConfigMap() (*corev1.ConfigMap, error) {
	// Create kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Get current namespace
	namespace, err := getCurrentNamespace()
	if err != nil {
		return nil, err
	}

	// Try to get existing configmap
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(
		context.TODO(),
		configMapName,
		metav1.GetOptions{},
	)

	if err == nil {
		return cm, nil
	}

	// Create new configmap if not exists
	newCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: map[string]string{},
	}

	cm, err = clientset.CoreV1().ConfigMaps(namespace).Create(
		context.TODO(),
		newCm,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func getCurrentNamespace() (string, error) {
	// Read namespace from pod
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getNextInConfigmap(configmap *corev1.ConfigMap, podSet string, bitsInID int) (string, error) {
	//Get current value or set default
	currentVal := defaultValue
	if val, exists := configmap.Data[podSet]; exists {
		currentVal = val
	}

	// Convert to int and increment
	counter, err := strconv.Atoi(currentVal)
	if err != nil {
		return "", fmt.Errorf("error converting counter to int: %w", err)
	}
	counter = (counter + 1) % (1 << bitsInID)

	// Update configmap with new value
	if configmap.Data == nil {
		configmap.Data = make(map[string]string)
	}
	configmap.Data[podSet] = strconv.Itoa(counter)

	// Create kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("error getting cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("error creating clientset: %w", err)
	}

	// Get current namespace
	namespace, err := getCurrentNamespace()
	if err != nil {
		return "", fmt.Errorf("error getting namespace: %w", err)
	}

	// Update configmap
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(
		context.TODO(),
		configmap,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("error updating configmap: %w", err)
	}

	return strconv.Itoa(counter), nil

}

func podDoesNotExist(label1Name, label1Value, label2Name, label2Value string) error {
	// Create kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error getting cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating clientset: %w", err)
	}

	// Create label selector string
	labelSelector := fmt.Sprintf("%s=%s,%s=%s",
		label1Name, label1Value,
		label2Name, label2Value)

	// List pods in all namespaces with labels
	pods, err := clientset.CoreV1().Pods("").List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: labelSelector,
		},
	)
	if err != nil {
		return fmt.Errorf("error listing pods: %w", err)
	}

	// Check if any pods found
	if len(pods.Items) > 0 {
		return fmt.Errorf("pod with labels %s already exists", labelSelector)
	}

	return nil
}
