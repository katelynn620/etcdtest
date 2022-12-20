# etcdtest

```
export ETCDCTL_API=3
export ETCDCTL_ENDPOINTS=http://127.0.0.1:2379

etcdctl member list

etcdctl --user root:1234 auth enable
etcdctl --user root:1234 user list
```