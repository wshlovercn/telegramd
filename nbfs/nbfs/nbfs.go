/*
 *  Copyright (c) 2018, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"github.com/golang/glog"

	upload "github.com/nebulaim/telegramd/nbfs/nbfs/upload/rpc"
	photo "github.com/nebulaim/telegramd/nbfs/nbfs/photo/rpc"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/BurntSushi/toml"
	"fmt"
	"github.com/nebulaim/telegramd/nbfs/biz/dal/dao"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"google.golang.org/grpc"
	"github.com/nebulaim/telegramd/nbfs/biz/core"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type RpcServerConfig struct {
	Addr string
}

type NbfsConfig struct {
	DataPath string
}

type uploadServerConfig struct {
	Nbfs      *NbfsConfig
	Server    *RpcServerConfig
	Discovery service_discovery.ServiceDiscoveryServerConfig
	Mysql     []mysql_client.MySQLConfig
}

// 整合各服务，方便开发调试
func main() {
	flag.Parse()

	config := &uploadServerConfig{}
	if _, err := toml.DecodeFile("./nbfs.toml", config); err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	glog.Info(config.Nbfs, ", ", config.Server, ", ", config.Discovery, ", ", config.Mysql)

	// Init
	core.InitNbfsDataPath(config.Nbfs.DataPath)

	// 初始化mysql_client、redis_client
	// redis_client.InstallRedisClientManager(config.Redis)
	mysql_client.InstallMysqlClientManager(config.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	// dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	// Start server
	grpcServer := grpc_util.NewRpcServer(config.Server.Addr, &config.Discovery)
	grpcServer.Serve(func(s *grpc.Server) {
		mtproto.RegisterRPCUploadServer(s, &upload.UploadServiceImpl{DataPath: config.Nbfs.DataPath})
		mtproto.RegisterRPCNbfsServer(s, &photo.PhotoServiceImpl{})
	})
}
