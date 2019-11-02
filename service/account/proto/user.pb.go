// Code generated by protoc-gen-go. DO NOT EDIT.
// source: user.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ReqSignup struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Passsword            string   `protobuf:"bytes,2,opt,name=passsword,proto3" json:"passsword,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqSignup) Reset()         { *m = ReqSignup{} }
func (m *ReqSignup) String() string { return proto.CompactTextString(m) }
func (*ReqSignup) ProtoMessage()    {}
func (*ReqSignup) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{0}
}

func (m *ReqSignup) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqSignup.Unmarshal(m, b)
}
func (m *ReqSignup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqSignup.Marshal(b, m, deterministic)
}
func (m *ReqSignup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqSignup.Merge(m, src)
}
func (m *ReqSignup) XXX_Size() int {
	return xxx_messageInfo_ReqSignup.Size(m)
}
func (m *ReqSignup) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqSignup.DiscardUnknown(m)
}

var xxx_messageInfo_ReqSignup proto.InternalMessageInfo

func (m *ReqSignup) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *ReqSignup) GetPasssword() string {
	if m != nil {
		return m.Passsword
	}
	return ""
}

type RespSignup struct {
	Code                 int32    `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RespSignup) Reset()         { *m = RespSignup{} }
func (m *RespSignup) String() string { return proto.CompactTextString(m) }
func (*RespSignup) ProtoMessage()    {}
func (*RespSignup) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{1}
}

func (m *RespSignup) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RespSignup.Unmarshal(m, b)
}
func (m *RespSignup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RespSignup.Marshal(b, m, deterministic)
}
func (m *RespSignup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RespSignup.Merge(m, src)
}
func (m *RespSignup) XXX_Size() int {
	return xxx_messageInfo_RespSignup.Size(m)
}
func (m *RespSignup) XXX_DiscardUnknown() {
	xxx_messageInfo_RespSignup.DiscardUnknown(m)
}

var xxx_messageInfo_RespSignup proto.InternalMessageInfo

func (m *RespSignup) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *RespSignup) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*ReqSignup)(nil), "proto.ReqSignup")
	proto.RegisterType((*RespSignup)(nil), "proto.RespSignup")
}

func init() { proto.RegisterFile("user.proto", fileDescriptor_116e343673f7ffaf) }

var fileDescriptor_116e343673f7ffaf = []byte{
	// 167 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0x2d, 0x4e, 0x2d,
	0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0xae, 0x5c, 0x9c, 0x41, 0xa9,
	0x85, 0xc1, 0x99, 0xe9, 0x79, 0xa5, 0x05, 0x42, 0x52, 0x5c, 0x1c, 0x20, 0x15, 0x79, 0x89, 0xb9,
	0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x70, 0xbe, 0x90, 0x0c, 0x17, 0x67, 0x41, 0x62,
	0x71, 0x71, 0x71, 0x79, 0x7e, 0x51, 0x8a, 0x04, 0x13, 0x58, 0x12, 0x21, 0xa0, 0x64, 0xc5, 0xc5,
	0x15, 0x94, 0x5a, 0x5c, 0x00, 0x35, 0x47, 0x88, 0x8b, 0x25, 0x39, 0x3f, 0x05, 0x62, 0x06, 0x6b,
	0x10, 0x98, 0x2d, 0x24, 0xc1, 0xc5, 0x9e, 0x9b, 0x5a, 0x5c, 0x9c, 0x98, 0x9e, 0x0a, 0xd5, 0x0d,
	0xe3, 0x1a, 0xd9, 0x71, 0x71, 0x87, 0x16, 0xa7, 0x16, 0x05, 0xa7, 0x16, 0x95, 0x65, 0x26, 0xa7,
	0x0a, 0xe9, 0x73, 0xb1, 0x41, 0x8d, 0x11, 0x80, 0x38, 0x55, 0x0f, 0xee, 0x40, 0x29, 0x41, 0xb8,
	0x08, 0xcc, 0x2e, 0x25, 0x86, 0x24, 0x36, 0xb0, 0x98, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xab,
	0xa5, 0xff, 0x5c, 0xde, 0x00, 0x00, 0x00,
}
