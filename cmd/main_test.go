package main

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	appsv1 "github.com/yasamprom/balancer-operator/api/v1"
	"github.com/yasamprom/balancer-operator/internal/controller"
	slicer "github.com/yasamprom/balancer-operator/internal/repo/clients/slicer"
	usecases "github.com/yasamprom/balancer-operator/internal/usecases"
	watcher "github.com/yasamprom/balancer-operator/internal/watcher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//+kubebuilder:scaffold:imports
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func BenchmarkCreation(b *testing.B) {
	os.Setenv("WATCH_PODS", "some-name")
	_ = startManager()
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func startManager() manager.Manager {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// creating clients
	c := slicer.New(slicer.Config{
		Host: slicerHost,
		Port: slicerPort,
	})

	webhookServer := webhook.NewServer(webhook.Options{})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f339fcfb.yasamprom.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.BalancerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Uc:     usecases.NewUsecases(c),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Balancer")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/stepan/.kube/config")
	if err != nil {
		panic(err)
	}

	// create k8s-api client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// create watcher
	triggers := parseArgs()
	if len(triggers.Labels) == 1 && len(triggers.Names) == 1 &&
		len(triggers.Labels[0]) == 0 && len(triggers.Names[0]) == 0 {
		log.Fatal("No watch rules set! Set filter rules.")
	}

	w, err := client.CoreV1().Pods("").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	watcher := watcher.New(watcher.Config{
		Watcher:  w,
		Triggers: triggers,
		Slicer:   c,
	})
	watcher.StartWatching(context.Background())

	setupLog.Info("Starting manager")
	// go func() {
	// 	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
	// 		setupLog.Error(err, "problem running manager")
	// 		os.Exit(1)
	// 	}
	// }()
	return mgr
}
