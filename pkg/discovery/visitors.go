package discovery

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/mhausenblas/reshifter/pkg/types"
	"golang.org/x/net/context"
)

// Visit2 recursively visits an etcd2 server from the path given and applies
// the reap function on leaf nodes, that is, keys that don't have sub-keys,
// otherwise descents the tree.
func Visit2(kapi client.KeysAPI, path string, reapfn types.Reap) error {
	log.WithFields(log.Fields{"func": "discovery.Visit2"}).Debug(fmt.Sprintf("Processing %s", path))
	res, err := kapi.Get(context.Background(),
		path,
		&client.GetOptions{
			Recursive: false,
			Quorum:    false,
		},
	)
	if err != nil {
		return err
	}
	if res.Node.Dir { // there are children, descent sub-tree
		log.WithFields(log.Fields{"func": "discovery.Visit2"}).Debug(fmt.Sprintf("%s has %d children", path, len(res.Node.Nodes)))
		for _, node := range res.Node.Nodes {
			log.WithFields(log.Fields{"func": "discovery.Visit2"}).Debug(fmt.Sprintf("Next I'm going to visit child %s", node.Key))
			_ = Visit2(kapi, node.Key, reapfn)
		}
		return nil
	}
	// we're on a leaf node, so apply the reap function:
	return reapfn(res.Node.Key, string(res.Node.Value))
}

// Visit3 visits the given path of an etcd3 server and applies the reap function
// on the keys in the respective range, depending on the Kubernetes distro.
func Visit3(c3 *clientv3.Client, path string, distro types.KubernetesDistro, reapfn types.Reap) error {
	log.WithFields(log.Fields{"func": "discovery.Visit3"}).Debug(fmt.Sprintf("Processing %s", path))
	endkey := ""
	if distro == types.Vanilla {
		endkey = types.KubernetesPrefixLast
	}
	if distro == types.OpenShift {
		endkey = types.OpenShiftPrefixLast
	}
	res, err := c3.Get(context.Background(), path+"/*", clientv3.WithRange(endkey))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"func": "discovery.Visit3"}).Debug(fmt.Sprintf("Got %v", res))
	for _, ev := range res.Kvs {
		log.WithFields(log.Fields{"func": "discovery.Visit3"}).Debug(fmt.Sprintf("key: %s, value: %s", ev.Key, ev.Value))
		err = reapfn(string(ev.Key), string(ev.Value))
		if err != nil {
			return err
		}
	}
	return nil
}