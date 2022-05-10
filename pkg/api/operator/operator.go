package operator

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	// controller "github.com/armory/spinnaker-operator/controllers"
	// "github.com/armory/spinnaker-operator/controllers/accountvalidating"
	"github.com/armory/spinnaker-operator/controllers/spinnakerservice"
	// "github.com/armory/spinnaker-operator/controllers/spinnakervalidating"
	"github.com/armory/spinnaker-operator/controllers/webhook"
	"github.com/armory/spinnaker-operator/pkg/api/version"

	// "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	// kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	// "github.com/operator-framework/operator-sdk/pkg/metrics"
	// sdkVersion "github.com/operator-framework/operator-sdk/version"			It is no longer possible to import package version

	"github.com/operator-framework/operator-lib/leader"
	ctrl "sigs.k8s.io/controller-runtime"

	// runtimehandler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	spinnakeriov1alpha2 "github.com/armory/spinnaker-operator/pkg/api/v1alpha2"
	"github.com/spf13/pflag"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme                    = kruntime.NewScheme()
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Spinnaker Operator Version: %v", version.GetOperatorVersion()))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	// log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
	// GetOperatorVersion
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(spinnakeriov1alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func Start(apiScheme func(s *kruntime.Scheme) error) {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	// pflag.CommandLine.AddFlagSet(zap.FlagSet())
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	fs := flag.FlagSet{}
	var disableAdmission bool

	defaultCertsDir := filepath.Join(getHome(), "spinnaker-operator-certs")
	fs.BoolVar(&disableAdmission, "disable-admission-controller", false, "Set to disable admission controller")
	fs.StringVar(&webhook.CertsDir, "certs-dir", defaultCertsDir, "Directory where tls.crt, tls.key and ca.crt files are found. Default: $HOME/spinnaker-operator-certs")
	pflag.CommandLine.AddGoFlagSet(&fs)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	// logf.SetLogger(zap.Logger())
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	printVersion()

	acc := spinnakerservice.TypesFactory.NewAccount()
	namespace := acc.GetNamespace()
	// namespace := "AAA-TMP"
	// namespace, _ := meta.NewAccessor().Namespace()
	// namespace, _ := meta.GetNamespace()
	// namespace, _ := k8sutil.GetWatchNamespace()
	if namespace != "" {
		log.Info(fmt.Sprintf("Watching Spinnaker configuration in %s", namespace))
	} else {
		log.Info("Watching Spinnaker configuration in cluster")
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "spinnaker-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ddafaf11.spinnaker.io",
		Namespace:              namespace,
		MapperProvider:         apiutil.NewDiscoveryRESTMapper,
		// MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apiScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Add admission controller
	// if !disableAdmission {
	// 	controller.Register(spinnakervalidating.Add)
	// 	controller.Register(accountvalidating.Add)
	// }

	// // Setup all Controllers
	// if err := controller.AddToManager(mgr); err != nil {
	// 	log.Error(err, "")
	// 	os.Exit(1)
	// }

	if !disableAdmission {
		log.Info("starting webhook server...")
		if err := webhook.Start(mgr); err != nil {
			log.Error(err, "error starting webhook server")
			os.Exit(1)
		}
	}

	// gvks, err := getGVKs(mgr)
	// if err != nil {
	// 	log.Error(err, "unable to get GroupVersionKind")
	// 	os.Exit(1)
	// }

	// runtimehandler.InstrumentedEnqueueRequestForObject
	// runtimehandler.EnqueueRequestForObject()
	// metrics := runtimehandler.EnqueueRequestForObject()
	// metrics.
	// Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {

	// Generates operator specific metrics based on the GVKs.
	// It serves those metrics on "http://metricsHost:operatorMetricsPort".
	// err = kubemetrics.GenerateAndServeCRMetrics(cfg, []string{namespace}, gvks, metricsHost, operatorMetricsPort)
	// if err != nil {
	// 	log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	// }

	// servicePorts := []v1.ServicePort{
	// 	{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	// 	{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
	// }
	// // Create Service object to expose the metrics port(s).
	// _, err = metrics.CreateMetricsService(ctx, cfg, servicePorts)
	// if err != nil {
	// 	log.Info(err.Error())
	// }

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func getGVKs(m manager.Manager) ([]schema.GroupVersionKind, error) {
	gvks := make([]schema.GroupVersionKind, 0)
	objs := []kruntime.Object{
		spinnakerservice.TypesFactory.NewService(),
		spinnakerservice.TypesFactory.NewAccount(),
	}
	for _, obj := range objs {
		gvk, err := apiutil.GVKForObject(obj, m.GetScheme())
		if err != nil {
			return nil, err
		}
		gvks = append(gvks, gvk)
	}
	return gvks, nil
}

func getHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	// Openshift will execute the container with random uid.
	if home == "/" {
		home = os.Getenv("OPERATOR_HOME")
	}
	return home
}
