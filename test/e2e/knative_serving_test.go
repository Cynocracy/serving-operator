package e2e

import (
	goctx "context"
	"testing"
	"time"

	"github.com/openshift-knative/knative-serving-operator/pkg/apis"
	servingv1alpha1 "github.com/openshift-knative/knative-serving-operator/pkg/apis/serving/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestKnativeServing(t *testing.T) {
	installList := &servingv1alpha1.InstallList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Install",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, installList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests in subgroups, maybe? we could re-organize this
	// however it makes sense.
	t.Run("knative-serving-group", func(t *testing.T) {
		t.Run("Cluster", KnativeServingCluster)
	})
}

func KnativeServingCluster(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "knative-serving-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	// create Install custom resource
	install := &servingv1alpha1.Install{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Install",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "knative-serving",
			Namespace: namespace,
		},
	}
	err = f.Client.Create(goctx.TODO(), install, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}
	deployments := []string{"controller", "activator", "autoscaler", "webhook"}
	for _, name := range deployments {
		err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, timeout)
		if err != nil {
			t.Fatal(err)
		}
	}
}
