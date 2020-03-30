# BANS 5GC Operator

A Kubernetes operator for deploying and managing BANS slices for BANS 5GC.

## Prerequisites

See [operator-framework/operator-sdk](https://github.com/operator-framework/operator-sdk#prerequisites).

```Shellsession
# Pre-install
sudo apt instal make

# Install dependencies
git clone -b bmv2-fabric https://github.com/stevenchiu30801/free5gc-operator.git
cd free5gc-operator && make install
git clone -b bmv2-fabric https://github.com/stevenchiu30801/onos-bandwidth-operator.git
cd onos-bandwidth-operator && make install
```

## Usage

### Run

```ShellSession
# Install all resources (CRD's, RBAC and Operator)
make install
```

### Procedure Test

```ShellSession
# Create a new CR
kubectl apply -f deploy/crds/bans.io_v1alpha1_bansslice_cr1.yaml

# Check if status of the new slice is ready before proceeding
kubectl describe bansslice.bans.io example-bansslice1 | grep -A 1 Status

# Set ransim pod variable
export RANSIM_POD=$( kubectl get pods -l app.kubernetes.io/instance=free5gc -l app.kubernetes.io/name=ransim -o jsonpath='{.items[0].metadata.name}' )

# Test registration and data traffic with slice 1
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=1 -sst=1 -sd=010203"

# Create a new CR
kubectl apply -f deploy/crds/bans.io_v1alpha1_bansslice_cr2.yaml

# Check if status of the new slice is ready before proceeding
kubectl describe bansslice.bans.io example-bansslice2 | grep -A 1 Status

# Test registration and data traffic with slice 2
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=2 -sst=1 -sd=112233"
```

### Reset

```ShellSession
# Uninstall all that all performed in the $ make install
make uninstall
```
