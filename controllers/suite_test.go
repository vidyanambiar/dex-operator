// // Copyright Red Hat

// package controllers

// import (
// 	"context"
// 	"path/filepath"
// 	"testing"

// 	"github.com/ghodss/yaml"
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes/scheme"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/envtest"
// 	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
// 	logf "sigs.k8s.io/controller-runtime/pkg/log"
// 	"sigs.k8s.io/controller-runtime/pkg/log/zap"

// 	dexoperatorconfig "github.com/identitatem/dex-operator/config"

// 	clusteradmasset "open-cluster-management.io/clusteradm/pkg/helpers/asset"

// 	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// 	authv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
// 	//+kubebuilder:scaffold:imports
// )

// // These tests use Ginkgo (BDD-style Go testing framework). Refer to
// // http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// var (
// 	k8sClient client.Client
// 	testEnv   *envtest.Environment
// 	ctx       context.Context
// 	cancel    context.CancelFunc
// )

// func TestAPIs(t *testing.T) {
// 	RegisterFailHandler(Fail)

// 	RunSpecsWithDefaultAndCustomReporters(t,
// 		"Controller Suite",
// 		[]Reporter{printer.NewlineReporter{}})
// }

// var _ = BeforeSuite(func() {
// 	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

// 	ctx, cancel = context.WithCancel(context.TODO())

// 	By("bootstrapping test environment")

// 	// Registering our APIs
// 	err := authv1alpha1.AddToScheme(scheme.Scheme)
// 	Expect(err).NotTo(HaveOccurred())

// 	//+kubebuilder:scaffold:scheme

// 	// Configure a new test environment which ingests our CRDs to allow an API server to know about our custom resources
// 	testEnv = &envtest.Environment{
// 		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
// 		ErrorIfCRDPathMissing: true,
// 	}

// 	// Start the environment (API server)
// 	cfg, err := testEnv.Start()
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(cfg).NotTo(BeNil())

// 	// Create a client to talk to the API server
// 	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(k8sClient).NotTo(BeNil())

// }, 60)

// var _ = AfterSuite(func() {
// 	By("tearing down the test environment")
// 	err := testEnv.Stop()
// 	Expect(err).NotTo(HaveOccurred())
// })

// var _ = Describe("Setup Dex", func() {
// 	By("Checking the CRDs availability", func() {
// 		// This is to test if the CRD are available through resources.go
// 		// as they are needed by other operators to dynamically install this operator
// 		readerDex := dexoperatorconfig.GetScenarioResourcesReader()
// 		_, err := getCRD(readerDex, "crd/bases/auth.identitatem.io_dexclients.yaml")
// 		Expect(err).Should(BeNil())

// 		_, err = getCRD(readerDex, "crd/bases/auth.identitatem.io_dexservers.yaml")
// 		Expect(err).Should(BeNil())
// 	})
// })

// var _ = Describe("Process DexServer:", func() {
// 	DexServerName := "my-dexserver"
// 	DexServerNamespace := "my-dexserver-ns"
// 	AuthRealmName := "my-authrealm"
// 	AuthRealmNameSpace := "my-authrealm-ns"
// 	MyGithubAppClientID := "my-github-app-client-id"
// 	DexServerIssuer := "https://testroutesubdomain.testhost.com"

// 	By("creating a DexServer CR", func() {
// 		// A DexServer object with metadata and spec.
// 		dexserver := &authv1alpha1.DexServer{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      DexServerName,
// 				Namespace: DexServerNamespace,
// 			},
// 			Spec: authv1alpha1.DexServerSpec{
// 				Issuer: DexServerIssuer,
// 				Connectors: []authv1alpha1.ConnectorSpec{
// 					{
// 						Name: "my-github",
// 						Type: "github",
// 						GitHub: authv1alpha1.GitHubConfigSpec{
// 							ClientID: MyGithubAppClientID,
// 							ClientSecretRef: corev1.SecretReference{
// 								Name:      AuthRealmName + "github",
// 								Namespace: AuthRealmNameSpace,
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}
// 		// ctx := context.Background()
// 		k8sClient.Create(context.TODO(), dexserver)
// 		// Expect(k8sClient.Create(context.TODO(), dexserver)).Should(Succeed())
// 	})
// })

// func getCRD(reader *clusteradmasset.ScenarioResourcesReader, file string) (*apiextensionsv1.CustomResourceDefinition, error) {
// 	b, err := reader.Asset(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	crd := &apiextensionsv1.CustomResourceDefinition{}
// 	if err := yaml.Unmarshal(b, crd); err != nil {
// 		return nil, err
// 	}
// 	return crd, nil
// }

package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	authv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = Describe("Process DexServer:", func() {
	DexServerName := "my-dexserver"
	DexServerNamespace := "my-dexserver-ns"
	AuthRealmName := "my-authrealm"
	AuthRealmNameSpace := "my-authrealm-ns"
	MyGithubAppClientID := "my-github-app-client-id"
	DexServerIssuer := "https://testroutesubdomain.testhost.com"

	By("creating a DexServer CR", func() {
		// A DexServer object with metadata and spec.
		dexserver := &authv1alpha1.DexServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DexServerName,
				Namespace: DexServerNamespace,
			},
			Spec: authv1alpha1.DexServerSpec{
				Issuer: DexServerIssuer,
				Connectors: []authv1alpha1.ConnectorSpec{
					{
						Name: "my-github",
						Type: "github",
						GitHub: authv1alpha1.GitHubConfigSpec{
							ClientID: MyGithubAppClientID,
							ClientSecretRef: corev1.SecretReference{
								Name:      AuthRealmName + "github",
								Namespace: AuthRealmNameSpace,
							},
						},
					},
				},
			},
		}

		// Objects to track in the fake client.
		objs := []runtime.Object{dexserver}

		// Register operator types with the runtime scheme.
		s := scheme.Scheme
		// err := authv1alpha1.AddToScheme(s)
		s.AddKnownTypes(authv1alpha1.GroupVersion, dexserver)
		// Expect(err).NotTo(HaveOccurred())

		// Create a fake client to mock API calls.
		cl := fake.NewFakeClientWithScheme(s, objs...)

		r := &DexServerReconciler{
			Client: cl,
			Scheme: s,
		}

		// mock request to simulate Reconcile() being called on an event for a watched resource
		req := reconcile.Request{}
		req.Name = DexServerName
		req.Namespace = DexServerNamespace

		_, err := r.Reconcile(context.TODO(), req)
		Expect(err).Should(BeNil())
	})
})
