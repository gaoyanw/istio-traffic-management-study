package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"encoding/json"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	v1 "k8s.io/api/authorization/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	checkHeader            = "x-ext-authz"
	allowedValue           = "allow"
	resultHeader           = "x-ext-authz-check-result"
	receivedResourceHeader = "x-ext-authz-received-x-goog-resources-plain"
	overrideHeader         = "x-ext-authz-additional-header-override"
	overrideGRPCValue      = "grpc-additional-header-override-value"
	resultAllowed          = "allowed"
	resultDenied           = "denied"
)

var (
	httpPort = flag.String("http", "8000", "HTTP server port")
	grpcPort = flag.String("grpc", "9000", "gRPC server port")
	denyBody = fmt.Sprintf("denied by ext_authz for not found header `%s: %s` in the request", checkHeader, allowedValue)
)

type (
	extAuthzServerV3 struct{}
)

// TODO: switch to use proto
type ResourceInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Permission string `json:"permission"`
	Container  string `json:"container"`
}

// ExtAuthzServer implements the ext_authz v2/v3 gRPC and HTTP check request API.
type ExtAuthzServer struct {
	grpcServer *grpc.Server
	httpServer *http.Server

	grpcV3 *extAuthzServerV3
	// For test only
	httpPort chan int
	grpcPort chan int
}

func (s *extAuthzServerV3) logRequest(allow string, request *authv3.CheckRequest) {
	httpAttrs := request.GetAttributes().GetRequest().GetHttp()
	log.Printf("[gRPCv3][%s]: %s%s, attributes: %v\n", allow, httpAttrs.GetHost(),
		httpAttrs.GetPath(),
		request.GetAttributes())
}

func (s *extAuthzServerV3) logRequest_debug(key string, value string) {
	log.Printf("%s: %v", key, value)

}
func (s *extAuthzServerV3) allow(request *authv3.CheckRequest) *authv3.CheckResponse {
	s.logRequest("allowed", request)
	return &authv3.CheckResponse{
		HttpResponse: &authv3.CheckResponse_OkResponse{
			OkResponse: &authv3.OkHttpResponse{},
		},
		Status: &status.Status{Code: int32(codes.OK)},
	}
}

func (s *extAuthzServerV3) deny(request *authv3.CheckRequest, errMsg string) *authv3.CheckResponse {
	s.logRequest("denied", request)
	return &authv3.CheckResponse{
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{Code: typev3.StatusCode_Forbidden},
				Body:   errMsg,
				Headers: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:   receivedResourceHeader,
							Value: getResoureInfo(request),
						},
					},
				},
			},
		},
		Status: &status.Status{Code: int32(codes.PermissionDenied)},
	}
}

func getResoureInfo(req *authv3.CheckRequest) string {
	headers := req.GetAttributes().GetRequest().GetHttp().GetHeaders()
	return headers["x-goog-resources-plain"]
}

// Check implements gRPC v3 check request.
func (s *extAuthzServerV3) Check(_ context.Context, request *authv3.CheckRequest) (*authv3.CheckResponse, error) {

	// 1. get the kubeconfig
	var config *rest.Config
	var err error
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", "~/.kube/config")
	}
	if err != nil {
		panic(err)
	}

	// 2. Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 3. extract user, and resource attributes from headers
	headers := request.GetAttributes().GetRequest().GetHttp().GetHeaders()
	log.Printf("receive headers, %v", headers)

	resourceInfoValue := getResoureInfo(request)
	var resourceInfo ResourceInfo
	err2 := json.Unmarshal([]byte(resourceInfoValue), &resourceInfo)
	if err2 != nil {
		return s.deny(request, fmt.Sprintf("Failed to unmarshal header value: %v", err)), nil
	}
	user := headers["user"]
	permissionSlice := strings.Split(resourceInfo.Permission, ".")
	permissionSlice_slash := strings.Split(resourceInfo.Permission, "/")
	resource := strings.Split(permissionSlice_slash[1], ".")[0]

	verb := permissionSlice[len(permissionSlice)-1]
	namespace := strings.Split(resourceInfo.Container, "/")[1]
	group := permissionSlice_slash[0]

	s.logRequest_debug("user", user)
	s.logRequest_debug("resource", resource)
	s.logRequest_debug("verb", verb)
	s.logRequest_debug("namespace", namespace)
	s.logRequest_debug("group: ", group)

	// 4. create SubjectAccessReview
	sar := &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			User: user,
			ResourceAttributes: &v1.ResourceAttributes{
				Resource:  resource,
				Verb:      verb,
				Namespace: namespace,
				Group:     group,
			},
		},
	}
	log.Printf("x-goog-resources-plain: %v", resourceInfoValue)
	log.Printf("SubjectAccessReview request: %v", sar)

	sarClient := clientset.AuthorizationV1().SubjectAccessReviews()
	response, err := sarClient.Create(context.Background(), sar, metav1.CreateOptions{})

	s.logRequest_debug("SubjectAccessReview.status.allowed", strconv.FormatBool(response.Status.Allowed))
	if err != nil {
		return s.deny(request, fmt.Sprintf("Failed to call SubjectAccessReview: %v", err)), nil
	}

	// 5. send response to envoy
	if response.Status.Allowed {
		return s.allow(request), nil
	}

	return s.deny(request, fmt.Sprintf("SubjectAccessReview failed\nx-goog-resources-plain:%s,\nSubjectAccessReviewSpec:%v", resourceInfoValue, sar.Spec)), nil
}

// ServeHTTP implements the HTTP check request.
func (s *ExtAuthzServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("[HTTP] read body failed: %v", err)
	}
	l := fmt.Sprintf("%s %s%s, headers: %v, body: [%s]\n", request.Method, request.Host, request.URL, request.Header, body)
	if allowedValue == request.Header.Get(checkHeader) {
		log.Printf("[HTTP][allowed]: %s", l)
		response.Header().Set(resultHeader, resultAllowed)
		response.Header().Set(overrideHeader, request.Header.Get(overrideHeader))
		response.Header().Set(receivedResourceHeader, l)
		response.WriteHeader(http.StatusOK)
	} else {
		log.Printf("[HTTP][denied]: %s", l)
		response.Header().Set(resultHeader, resultDenied)
		response.Header().Set(overrideHeader, request.Header.Get(overrideHeader))
		response.Header().Set(receivedResourceHeader, l)
		response.WriteHeader(http.StatusForbidden)
		_, _ = response.Write([]byte(denyBody))
	}
}

func (s *ExtAuthzServer) startGRPC(address string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Printf("Stopped gRPC server")
	}()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
		return
	}
	// Store the port for test only.
	s.grpcPort <- listener.Addr().(*net.TCPAddr).Port

	s.grpcServer = grpc.NewServer()
	authv3.RegisterAuthorizationServer(s.grpcServer, s.grpcV3)

	log.Printf("Starting gRPC server at %s", listener.Addr())
	if err := s.grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
		return
	}
}

func (s *ExtAuthzServer) startHTTP(address string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Printf("Stopped HTTP server")
	}()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}
	// Store the port for test only.
	s.httpPort <- listener.Addr().(*net.TCPAddr).Port
	s.httpServer = &http.Server{Handler: s}

	log.Printf("Starting HTTP server at %s", listener.Addr())
	if err := s.httpServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (s *ExtAuthzServer) run(httpAddr, grpcAddr string) {
	var wg sync.WaitGroup
	wg.Add(2)
	go s.startHTTP(httpAddr, &wg)
	go s.startGRPC(grpcAddr, &wg)
	wg.Wait()
}

func (s *ExtAuthzServer) stop() {
	s.grpcServer.Stop()
	log.Printf("GRPC server stopped")
	log.Printf("HTTP server stopped: %v", s.httpServer.Close())
}

func NewExtAuthzServer() *ExtAuthzServer {
	return &ExtAuthzServer{
		grpcV3:   &extAuthzServerV3{},
		httpPort: make(chan int, 1),
		grpcPort: make(chan int, 1),
	}
}

func main() {
	flag.Parse()
	s := NewExtAuthzServer()
	go s.run(fmt.Sprintf(":%s", *httpPort), fmt.Sprintf(":%s", *grpcPort))
	defer s.stop()

	// Wait for the process to be shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
