# Balancer-operator
This is an operator for deploying custom load balancer.
Operator works with Slicer library and notifies it about connected/disconnected pods.
Second responsability of operator is deploying balancer which also communicates with Slicer.

## Before you start

Make sure you have configured k8s cluster.
1. For local development you can use [minikube](https://minikube.sigs.k8s.io/docs/)
2. For building balancer and custom app you can use [docker-desktop](https://www.docker.com/products/docker-desktop/)


## Description
Operator is k8s object, which can handle different events like pod connections or other.
Handling events of your pods can work by knowing its name pattern or special label.
Both ways are supported.

If your pods have names `awesome-pod-...` you may set
```
export WATCH_PODS=awesome-pod // common prefix
```
If you want to use only specific pods, use full names with `;` separetor
```
export WATCH_PODS=`awesome-pod-f32yu2;awesome-pod-wr81jh` // exact pods
```

## Run client app
Firstly, make docker images available for your cluster
```
eval $(minikube -p minikube docker-env)  
```
Then check available images
```
minikube image ls --format table
```
Build your image
```
docker build -t simple:v1 .
minikube cache add balancer:v1
```
Run image
```
kubectl run simple-v1  --image=simple:v1 --image-pull-policy=IfNotPresent
```

You don't have to run images, instead of this you can use any k8s objects.


## Operator developing/testing 

### Prerequisites
- go version v1.20.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/balancer-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/balancer-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

