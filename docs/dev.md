# Development

If you plan to contribute to ReShifter, please read the [contributor guidelines](../CONTRIBUTING.md) first and note the below.

## Builds and releases

ReShifter is developed in [Go 1.8](https://beta.golang.org/doc/go1.8), so please make sure that you've got at least this version installed.
We're following [semantic versioning](http://semver.org/). The canonical ReShifter release version is defined in one place only,
in the first line of the [Makefile](https://github.com/mhausenblas/reshifter/blob/master/Makefile), following semver, for example, `0.3.14`.

This version is then used in the Go code, in the Docker image as a tag and for all downstream deployments.

A new release (Linux binary on GitHub and image on quay.io) is cut using the following process:

```
# 0. Make sure all tests pass:
$ make gtest

# 1. Generate the binary:
$ make gbuild

# 2. Release on GitHub, using `v$reshifter_version`

# 3. Build a container image locally and push it to quay.io:
$ make crelease
```

## Vendoring

We are using Go [dep](https://github.com/golang/dep) for dependency management.
If you don't have `dep` installed yet, do `go get -u github.com/golang/dep/cmd/dep` now and then:

```
$ dep ensure
```

## Unit tests

In general, for unit tests we use the `go test` command, for example:

```
$ go test -v -short -run Test* ./pkg/discovery
```

Please do make sure all unit tests pass before sending in a PR. Also, note that we apply [CAT](https://medium.com/@mhausenblas/container-assisted-testing-b76ee74278b7), so in order for the unit tests to run you need to have Docker running.

## Local testing

The following shows an example (interactive) session against an etcd3-based Kubernetes control plane.

First, launch ReShifter:

```
$ docker run --rm -e "ACCESS_KEY_ID=Q3AM3UQ867SPQQA43P2F" -e "SECRET_ACCESS_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG" --name reshifter -p 8080:8080 quay.io/mhausenblas/reshifter:0.3.7
```

Now, launch a local etcd:

```
$ docker run --rm -p 2379:2379 --name test-etcd --dns 8.8.8.8 quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd  \
--advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380
```

Note: use the result of the following command as the value for the etcd endpoint in the ReShifter UI/API:

```
$ docker inspect test-etcd | jq -r '.[0].NetworkSettings.IPAddress'
```

Next we generate some entries in etcd:

```
$ export ETCDCTL_API=3
$ etcdctl --endpoints=http://127.0.0.1:2379 put /kubernetes.io "."
$ etcdctl --endpoints=http://127.0.0.1:2379 put /kubernetes.io/namespaces/kube-system "."
$ etcdctl --endpoints=http://127.0.0.1:2379 put /openshift.io "."
```

Now you can use the UI to create a backup and after restarting etcd3 you can restore it again.

You can also upload an existing backup file via the CLI like this:

```
$ curl --form 'backupfile=@/tmp/test/149958881.zip' http://localhost:8080/v1/restore/upload
```

Note that if you want to use etc2, do the following:

```
$ docker run --rm -p 2379:2379 --name test-etcd --dns 8.8.8.8 quay.io/coreos/etcd:v2.3.8 --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379
$ curl http://127.0.0.1:2379/v2/keys/kubernetes.io/namespaces/kube-system -XPUT -d value="."
$ curl http://127.0.0.1:2379/v2/keys/openshift.io -XPUT -d value="."
```

## Scheduled backups

```
$ kubectl create -f https://raw.githubusercontent.com/mhausenblas/reshifter/master/deployments/hourly-backup-job.yaml

$ kubectl get cronjob
NAME            SCHEDULE    SUSPEND   ACTIVE    LAST-SCHEDULE
hourly-backup   0 * * * *   False     0         Thu, 06 Jul 2017 07:00:00 +0100

$ kubectl delete cronjob/hourly-backup
cronjob "hourly-backup" deleted
```

## Demo

The demo given to the Kubernetes SIG Cluster Lifecycle on 2017-06-27:

```
# Use Minio playground as S3 compatible storage backend:
cd /Users/mhausenblas/Dropbox/dev/work/src/github.com/mhausenblas/reshifter
export ACCESS_KEY_ID=Q3AM3UQ867SPQQA43P2F
export SECRET_ACCESS_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

# Launch etcd2:
docker run --rm -p 2379:2379 \
           --name test-etcd --dns 8.8.8.8 quay.io/coreos/etcd:v2.3.8 \
          --advertise-client-urls http://0.0.0.0:2379 \
          --listen-client-urls http://0.0.0.0:2379

# Launch ReShifter:
cd /Users/mhausenblas/Dropbox/dev/work/src/github.com/mhausenblas/reshifter
DEBUG=true reshifter

# Populate etcd:
curl http://localhost:2379/v2/keys/kubernetes.io/namespaces/kube-system -XPUT -d \
         value="{\"kind\":\"Namespace\",\"apiVersion\":\"v1\"}"

# Backup via UI:
# Open http://localhost:8080/reshifter/
# Config -> Backup
# Open https://play.minio.io:9000/ and verify with bucket

# Re-start etcd:
docker kill test-etcd
docker run --rm -p 2379:2379 \
           --name test-etcd --dns 8.8.8.8 quay.io/coreos/etcd:v2.3.8 \
          --advertise-client-urls http://0.0.0.0:2379 \
          --listen-client-urls http://0.0.0.0:2379

# Restore via UI:
# Open http://localhost:8080/reshifter/
# Restore

# Query etcd to verify restore:
http http://localhost:2379/v2/keys/kubernetes.io/namespaces/kube-system
```

## etcd key prefixes

### Kubernetes

```
/kubernetes.io/ranges
/kubernetes.io/statefulsets
/kubernetes.io/jobs
/kubernetes.io/horizontalpodautoscalers
/kubernetes.io/events
/kubernetes.io/masterleases
/kubernetes.io/minions
/kubernetes.io/persistentvolumes
/kubernetes.io/configmaps
/kubernetes.io/controllers
/kubernetes.io/deployments
/kubernetes.io/serviceaccounts
/kubernetes.io/services
/kubernetes.io/namespaces
/kubernetes.io/securitycontextconstraints
/kubernetes.io/thirdpartyresources
/kubernetes.io/persistentvolumeclaims
/kubernetes.io/pods
/kubernetes.io/replicasets
/kubernetes.io/secrets
```

### OpenShift

```
/openshift.io/authorization
/openshift.io/buildconfigs
/openshift.io/oauth
/openshift.io/registry
/openshift.io/users
/openshift.io/useridentities
/openshift.io/builds
/openshift.io/deploymentconfigs
/openshift.io/images
/openshift.io/imagestreams
/openshift.io/ranges
/openshift.io/routes
/openshift.io/templates
```
