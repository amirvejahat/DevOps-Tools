package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("Error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}
	namespace := flag.String("namespace", "default", "select namespace")
	flag.Parse()
	persistentUsedVolumeList := []string{}
	pods, err := ListPods(*namespace, clientset)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, pod := range pods.Items {
		for _, v := range pod.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				persistentUsedVolumeList = append(persistentUsedVolumeList, v.PersistentVolumeClaim.ClaimName)
			}
		}
	}

	allVolumeClaimList, err := ListPersistentVolumeClaims(*namespace, clientset)
	if err != nil {
		fmt.Printf("Error getting persistentVolumeClaims: %v\n", err)
		os.Exit(1)
	}
	for _, pvc := range allVolumeClaimList {
		if contains(persistentUsedVolumeList, pvc.Name) {
			continue
		}
		fmt.Printf("PersistentVolumeClaim = %v is unused.\n", pvc.Name)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func ListPods(namespace string, client kubernetes.Interface) (*v1.PodList, error) {
	fmt.Println("Get Kubernetes Pods")
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("error getting pods: %v", err)
		return nil, err
	}
	return pods, nil
}

func ListPersistentVolumeClaims(namespace string, client kubernetes.Interface) (v []v1.PersistentVolumeClaim, err error) {
	list, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pvc := range list.Items {
		// List of all persistent volumes
		v = append(v, pvc)
	}
	return v, nil
}
