package main

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	labelSet            = "danieldorado.github.io/podtinytidyid-set"
	labelId             = "danieldorado.github.io/podtinytidyid-id"
	varId               = "PODTINYTIDYID_ID"
	containers          = "containers"
	initContainers      = "initContainers"
	ephemeralContainers = "ephemeralContainers"
	patchLabels         = `{"op": "add", "path": "/metadata/labels/%s", "value": "%s"}`
	// Fields: containerType, containerIndex, name, value
	patchEnvNew    = `{"op": "add", "path": "/spec/%s/%d/env", "value": [{"name": "%s", "value": "%s"}]}`
	patchEnvAppend = `{"op": "add", "path": "/spec/%s/%d/env/-", "value": {"name": "%s", "value": "%s"}}`
)

func getLabelId(podSet string) string {
	return labelId + "-" + podSet
}

func (c *Config) mutatePods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.Info("mutatePods")
	// check inputs
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource { // check resource
		err := fmt.Errorf("wrong Resource received in pod. Expected: %+v. Get: %+v",
			podResource, ar.Request.Resource)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	podKind := metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
	if ar.Request.Kind != podKind { // check kind
		err := fmt.Errorf("wrong Kind received in mutatePods. Expected: %+v. Get: %+v",
			podKind, ar.Request.Kind)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	// deserialize
	pod := corev1.Pod{}
	decoder := codecs.UniversalDeserializer()

	_, _, err := decoder.Decode(ar.Request.Object.Raw, nil, &pod)
	if err != nil {
		err = fmt.Errorf("decoding in mutatePods: %w", err)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	podSet := pod.ObjectMeta.Labels[labelSet]
	podId := GetNext(podSet, c.Generator.IdentifierBits)

	// Crete patch
	labelsToInject := map[string]string{getLabelId(podSet): podId}
	envarsToInject := map[string]string{varId: podId}
	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true
	patches := []string{}
	for k, v := range labelsToInject {
		kScaped := strings.Replace(k, "/", "~1", -1)
		klog.Infof("mutatePods: injecting label %s with value %s", kScaped, v)
		patches = append(patches, fmt.Sprintf(patchLabels, kScaped, v))
	}
	for k, v := range envarsToInject {
		// Handle containers
		klog.Infof("mutatePods: injecting envar %s with value %s", k, v)
		for idx, container := range pod.Spec.Containers {
			if len(container.Env) == 0 {
				patches = append(patches, fmt.Sprintf(patchEnvNew, containers, idx, k, v))
			} else {
				patches = append(patches, fmt.Sprintf(patchEnvAppend, containers, idx, k, v))
			}
		}

		// Handle init containers
		for idx, container := range pod.Spec.InitContainers {
			if len(container.Env) == 0 {
				patches = append(patches, fmt.Sprintf(patchEnvNew, initContainers, idx, k, v))
			} else {
				patches = append(patches, fmt.Sprintf(patchEnvAppend, initContainers, idx, k, v))
			}
		}

		// Handle ephemeral containers
		for idx, container := range pod.Spec.EphemeralContainers {
			if len(container.EphemeralContainerCommon.Env) == 0 {
				patches = append(patches, fmt.Sprintf(patchEnvNew, ephemeralContainers, idx, k, v))
			} else {
				patches = append(patches, fmt.Sprintf(patchEnvAppend, ephemeralContainers, idx, k, v))
			}
		}
	}

	reviewResponse.Patch = []byte("[\n" + strings.Join(patches, ",\n") + "\n]")
	patchType := v1.PatchTypeJSONPatch
	reviewResponse.PatchType = &patchType
	klog.Infof("mutatePods applying patch in pod %s/%s?????:\n%s", pod.Namespace, pod.GenerateName, reviewResponse.Patch)
	return &reviewResponse
}
