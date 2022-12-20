package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var requestTimeout time.Duration = time.Second * 3

func main() {

	cli := getCli()

	Put(cli)

	Get(cli, "sample_key")

	auth := Auth{cli}

	auth.NormalUser()

	// auth.AddUser("test", "1234")

	// auth.AddRole("testrole")

	// auth.GrantPermission("testrole", "test_key", clientv3.GetPrefixRangeEnd("test_key"), clientv3.PermissionType(clientv3.PermReadWrite))
}

func getCli() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
		Username:    "root",
		Password:    "1234",
	})
	if err != nil {
		panic("cannot connect to etcd")
	}

	return cli
}

func Put(cli *clientv3.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Put(ctx, "sample_key", "sample_value")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("current revision:", resp.Header.Revision) // revision start at 1
	// current revision: 2
}

func Get(cli *clientv3.Client, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, key)
	// get with prefix
	//resp, err := cli.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
	// foo : bar
}

func (a *Auth) NormalUser() {
	a.AddUser("user1", "123")
	a.AddRole("role1")
	a.UserBindRole("user1", "role1")

	a.GrantPermission("role1", "user1", clientv3.GetPrefixRangeEnd("user1"), clientv3.PermissionType(clientv3.PermReadWrite))

	userCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
		Username:    "user1",
		Password:    "123",
	})
	if err != nil {
		log.Fatal(err)
	}
	userCli.KV.Put(context.TODO(), "user1/123", "value")

	resp, err := userCli.KV.Get(context.TODO(), "user3/123")
	if err != nil {
		// fmt.Print("error\n")
		log.Fatal(err)
	} else {
		fmt.Printf("get resp %v\n", resp)
		for _, ev := range resp.Kvs {
			fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		}
	}

	a.RoleDelete("role1")

	a.UserDelete("user1")
}

type Auth struct {
	Cli *clientv3.Client
}

func (a *Auth) AddUser(user, passwd string) {
	if _, err := a.Cli.UserAdd(context.TODO(), user, passwd); err != nil {
		if !a.IsAlreadyExists(err) {
			log.Fatal("UserAdd ", err)
		}
	}
}

func (a *Auth) IsAlreadyExists(err error) bool {
	return strings.Contains(fmt.Sprintf("%s", err), "already exists")
}

// 添加角色
func (a *Auth) AddRole(role string) {
	if _, err := a.Cli.RoleAdd(context.TODO(), role); err != nil {
		if !a.IsAlreadyExists(err) {
			log.Fatal("RoleAdd ", err)
		}
	}
}

func (a *Auth) UserBindRole(user, role string) {
	if _, err := a.Cli.UserGrantRole(context.TODO(), user, role); err != nil {
		if !a.IsAlreadyExists(err) {
			log.Fatal("UserGrantRole ", err)
		}
	}
}

func (a *Auth) GrantPermission(user, key, rangeEnd string, permission clientv3.PermissionType) {
	if resp, err := a.Cli.RoleGrantPermission(
		context.TODO(),
		user,     // role name
		key,      // key
		rangeEnd, // range end
		permission,
	); err != nil {
		log.Fatal("RoleGrantPermission ", err)
	} else {
		fmt.Printf("RoleGrantPermission resp %v\n", resp)
	}
}

func (a *Auth) RoleDelete(role string) {
	if _, err := a.Cli.RoleDelete(
		context.TODO(),
		role,
	); err != nil {
		log.Fatal("RoleDelete ", err)
	} else {
		fmt.Printf("Role %v was deleted.\n", role)
	}
}

func (a *Auth) UserDelete(user string) {
	if _, err := a.Cli.UserDelete(
		context.TODO(),
		user,
	); err != nil {
		log.Fatal("UserDelete ", err)
	} else {
		fmt.Printf("User %v was deleted.\n", user)
	}
}
