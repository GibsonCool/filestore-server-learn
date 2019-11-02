package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/service/account/proto"
	"filestore-server/util"
	"log"
)

type User struct {
}

func (*User) Signup(ctx context.Context, req *proto.ReqSignup, resp *proto.RespSignup) error {
	username := req.Username
	passwd := req.Passsword

	log.Printf("接受到来自rpc的请求，username:%s  pwd:%s", username, passwd)

	if len(username) < 3 || len(passwd) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "Invalid parameter: 用户名或密码不符合规范"
		return nil
	}

	//3.用户密码加盐处理
	encPasswd := util.Sha1([]byte(config.PwdSalt + passwd))
	//4.存入数据库 tbl_user 表并返回结果
	isSuccess := dblayer.UserSignUp(username, encPasswd)

	if isSuccess {
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	} else {
		resp.Code = common.StatusRegisterFailed
		resp.Message = "注册失败"
	}
	return nil
}
