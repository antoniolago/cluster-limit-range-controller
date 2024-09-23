# cluster-limit-range-controller

## Overview

The **cluster-limit-range-controller** is a Kubernetes custom controller that manages `LimitRange` objects across all namespaces based on a defined `ClusterLimitRange` custom resource. This controller ensures that specified resource limits are applied only to selected namespaces while respecting user-defined exceptions.

## Features

- Automatically applies `LimitRange` objects to all or specific namespaces.
- Ignores namespaces defined in the `ignoredNamespaces` list.
- Sync `LimitRange` objects when updated.

## Getting Started
```yaml
apiVersion: lag0.com.br/v1
kind: ClusterLimitRange
metadata:
  name: default-cluster-lr
spec:
# List of namespaces to be ignored
  ignoredNamespaces:
    - kube-system        
    - kube-public
  # List of namespaces to apply LimitRange, leave it empty to apply to all (except if ignored)
  applyNamespaces:
    - dev                
    - staging
  limits:
    - type: Container     # Type of resource limits
      max:
        cpu: "2"         # Maximum CPU limit
        memory: "4Gi"    # Maximum memory limit
      min:
        cpu: "200m"      # Minimum CPU limit
        memory: "256Mi"   # Minimum memory limit
      default:
        cpu: "500m"      # Default CPU limit
        memory: "1Gi"    # Default memory limit
      defaultRequest:
        cpu: "300m"      # Default request CPU limit
        memory: "512Mi"   # Default request memory limit
      maxLimitRequestRatio:
        cpu: "4"         # Max limit request ratio for CPU
```

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/cluster-limit-range-controller:tag
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
make deploy IMG=<some-registry>/cluster-limit-range-controller:tag
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

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/cluster-limit-range-controller:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/cluster-limit-range-controller/<tag or branch>/dist/install.yaml
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

