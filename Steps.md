1. root@k8clusters31:~/Kleaner# operator-sdk init --domain linuxdatahub.com --repo github.com/babugeet/test-webserver-operator
INFO[0000] Writing kustomize manifests for you to edit... 
INFO[0000] Writing scaffold for you to edit...          
INFO[0000] Get controller runtime:
$ go get sigs.k8s.io/controller-runtime@v0.16.3 
INFO[0003] Update dependencies:
$ go mod tidy           
Next: define a resource with:
$ operator-sdk create api


2. 

#Note: Kind should start with uppercase letter

root@k8clusters31:~/Kleaner# operator-sdk create api --group ldhctlr  --version v1alpha1 --kind Webserver 
INFO[0000] Create Resource [y/n]                        
y
INFO[0002] Create Controller [y/n]                      
y
INFO[0004] Writing kustomize manifests for you to edit... 
INFO[0004] Writing scaffold for you to edit...          
INFO[0004] api/v1alpha1/webserver_types.go              
INFO[0004] api/v1alpha1/groupversion_info.go            
INFO[0004] internal/controller/suite_test.go            
INFO[0004] internal/controller/webserver_controller.go  
INFO[0004] internal/controller/webserver_controller_test.go 
INFO[0004] Update dependencies:
$ go mod tidy           
INFO[0004] Running make:
$ make generate                
mkdir -p /root/Kleaner/bin
test -s /root/Kleaner/bin/controller-gen && /root/Kleaner/bin/controller-gen --version | grep -q v0.13.0 || \
GOBIN=/root/Kleaner/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
/root/Kleaner/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
Next: implement your new API and generate the manifests (e.g. CRDs,CRs) with:
$ make manifests
root@k8clusters31:~/Kleaner# 


root@k8clusters31:~/Kleaner# make generate
/root/Kleaner/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
root@k8clusters31:~/Kleaner# 

root@k8clusters31:~/Kleaner# make manifests
/root/Kleaner/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
root@k8clusters31:~/Kleaner# 