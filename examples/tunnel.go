package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/khulnasoft-lab/tunnel-kubernetes/pkg/artifacts"
	"github.com/khulnasoft-lab/tunnel-kubernetes/pkg/k8s"
	"github.com/khulnasoft-lab/tunnel-kubernetes/pkg/tunnelk8s"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"context"
)

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx := context.Background()

	cluster, err := k8s.GetCluster()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current namespace:", cluster.GetCurrentNamespace())

	tunnelk8sCopy := tunnelk8s.New(cluster, logger.Sugar(), tunnelk8s.WithExcludeOwned(true))
	tunnelk8s := tunnelk8s.New(cluster, logger.Sugar(), tunnelk8s.WithExcludeOwned(true))

	fmt.Println("Scanning cluster")

	//tunnel k8s #cluster
	artifacts, err := tunnelk8s.ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning kind 'pods' with exclude-owned=true")
	artifacts, err = tunnelk8s.Resources("pod").AllNamespaces().ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning namespace 'default'")
	//tunnel k8s --namespace default
	artifacts, err = tunnelk8sCopy.Namespace("default").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)
	fmt.Println("Scanning all namespaces ")
	artifacts, err = tunnelk8sCopy.AllNamespaces().ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning namespace 'default', resource 'deployment/orion'")

	//tunnel k8s --namespace default deployment/orion
	artifact, err := tunnelk8sCopy.Namespace("default").GetArtifact(ctx, "deploy", "orion")
	if err != nil {
		log.Fatal(err)
	}
	printArtifact(artifact)

	fmt.Println("Scanning 'deployments'")

	//tunnel k8s deployment
	artifacts, err = tunnelk8sCopy.Namespace("default").Resources("deployment").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	fmt.Println("Scanning 'cm,pods'")
	//tunnel k8s clusterroles,pods
	artifacts, err = tunnelk8sCopy.Namespace("default").Resources("cm,pods").ListArtifacts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printArtifacts(artifacts)

	tolerations := []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOperator(corev1.NodeSelectorOpExists),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: pointer.Int64(300),
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOperator(corev1.NodeSelectorOpExists),
			TolerationSeconds: pointer.Int64(300),
		},
	}

	// collect node info
	ar, err := tunnelk8sCopy.ListArtifactAndNodeInfo(ctx, "tunnel-temp", map[string]string{"chen": "test"}, tolerations...)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range ar {
		if a.Kind != "NodeInfo" {
			continue
		}
		fmt.Println(a.RawResource)
	}

	bi, err := tunnelk8s.ListBomInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bb, err := json.Marshal(bi)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(bb))
}

func printArtifacts(artifacts []*artifacts.Artifact) {
	for _, artifact := range artifacts {
		printArtifact(artifact)
	}
}

func printArtifact(artifact *artifacts.Artifact) {
	fmt.Printf(
		"Name: %s, Kind: %s, Namespace: %s, Images: %v\n",
		artifact.Name,
		artifact.Kind,
		artifact.Namespace,
		artifact.Images,
	)
}
