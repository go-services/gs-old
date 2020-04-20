package service

import (
	"fmt"
	"github.com/go-services/code"
	"github.com/ozgio/strutil"
	"strings"
)

var goToProtoTypeMap = map[string]string{
	"float64": "double",
	"float32": "float",
	"int32":   "int32",
	"int64":   "int64",
	"int":     "int64",
	"uint32":  "uint32",
	"uint64":  "uint64",
	"bool":    "bool",
	"string":  "string",
}

type ProtoMessage struct {
	Name   string
	Params []ProtoMessageParam
	Type   code.Type
	Struct *code.Struct
}
type ProtoMessageParam struct {
	Repeat   bool
	Name     string
	Type     string
	GoType   code.Type
	Position int
	Message  *ProtoMessage
}
type GRPCEndpoint struct {
	Name            string
	Endpoint        Endpoint
	RequestMessage  ProtoMessage
	ResponseMessage ProtoMessage
	Messages        []ProtoMessage
}

type GRPCTransport struct {
	GRPCEndpoint []GRPCEndpoint
}

func (p *ProtoMessageParam) String() string {
	s := ""
	if p.Repeat {
		s += "repeated"
	}
	return fmt.Sprintf("%s %s %s = %d", s, p.Type, p.Name, p.Position)
}

func (m *ProtoMessage) String() string {
	s := fmt.Sprintf("message %s {\n", m.Name)
	for _, v := range m.Params {
		s += v.String() + ";\n"
	}
	s += "}\n"
	return s
}

func parseGRPCTransport(svc Service) *GRPCTransport {
	tp := &GRPCTransport{}
	seen := map[string]*ProtoMessage{}
	for _, ep := range svc.Endpoints {
		grpcAnnotations := findAnnotations("grpc", ep.Annotations)
		if len(grpcAnnotations) == 0 {
			continue
		}
		tp.GRPCEndpoint = append(tp.GRPCEndpoint, parseGRPCEndpoint(ep, seen))
	}
	if len(tp.GRPCEndpoint) == 0 {
		return nil
	}
	return tp
}
func parseGRPCEndpoint(ep Endpoint, seen map[string]*ProtoMessage) GRPCEndpoint {
	grpcEp := GRPCEndpoint{
		Name:     ep.Name,
		Endpoint: ep,
	}
	if ep.Request != nil {
		message := generateMessage(&grpcEp.Messages, ep.Params[1].Type, ep.Request, seen)
		grpcEp.RequestMessage = message
	} else {
		var empty ProtoMessage
		if e, ok := seen["Empty"]; ok {
			empty = *e
		} else {
			empty = ProtoMessage{
				Name: "Empty",
			}
			seen["Empty"] = &empty
			grpcEp.Messages = append(grpcEp.Messages, empty)
		}
		grpcEp.RequestMessage = empty
	}
	responseMessage := ProtoMessage{
		Name: ep.Name + "Response",
		Params: []ProtoMessageParam{
			{
				Repeat:   false,
				Name:     "Err",
				Type:     "string",
				GoType:   code.NewType("string"),
				Position: 1,
			},
		},
	}
	if ep.Response != nil {
		message := generateMessage(&grpcEp.Messages, ep.Results[0].Type, ep.Response, seen)
		responseMessage.Params = append(responseMessage.Params, ProtoMessageParam{
			Repeat:   false,
			Name:     "Response",
			Message:  &message,
			GoType:   message.Type,
			Type:     getMessageName(ep.Results[0].Type.Import, ep.Response.Name),
			Position: 2,
		})

	}
	grpcEp.ResponseMessage = responseMessage
	grpcEp.Messages = append(grpcEp.Messages, responseMessage)
	return grpcEp
}

func getMessageName(imp *code.Import, name string) string {
	return strings.Title(strutil.ToCamelCase(imp.Alias)) + name
}

func generateMessage(messages *[]ProtoMessage, tp code.Type, structure *code.Struct, seen map[string]*ProtoMessage) ProtoMessage {
	name := getMessageName(tp.Import, structure.Name)
	if message, ok := seen[name]; ok {
		return *message
	}
	message := ProtoMessage{
		Struct: structure,
		Type:   tp,
		Name:   name,
	}
	seen[name] = &message
	for _, field := range structure.Fields {
		tag := ""
		if field.Tags != nil {
			tag = getTag("grpc", *field.Tags)
		}
		if !isExported(field.Name) || tag == "-" {
			continue
		}
		param := ProtoMessageParam{
			Repeat:   field.Type.ArrayType,
			Name:     field.Name,
			GoType:   field.Type,
			Position: len(message.Params) + 1,
		}
		if protoType, ok := goToProtoTypeMap[field.Type.Qualifier]; ok {
			param.Type = protoType
			message.Params = append(message.Params, param)
			continue
		}
		if field.Type.String() == "[]byte" {
			param.Type = "bytes"
			message.Params = append(message.Params, param)
			continue
		}
		// if the type is not exported we need to ignore it
		if !isExported(field.Type.Qualifier) {
			continue
		}

		// recursive struct
		if field.Type.Qualifier == structure.Name && field.Type.Import == nil {
			param.Type = message.Name
			message.Params = append(message.Params, param)
			continue
		}
		if field.Type.Import != nil {
			msgName := getMessageName(field.Type.Import, field.Type.Qualifier)
			obj, ok := seen[msgName]
			if !ok {
				s, err := findStruct(field.Type)
				if err != nil {
					log.Warnf("Type %s not supported\n", field.Type)
					continue
				}
				newMessage := generateMessage(messages, field.Type, s, seen)
				obj = &newMessage
			}
			param.Type = msgName
			param.Message = obj
			message.Params = append(message.Params, param)
			continue
		}
	}
	*messages = append(*messages, message)
	return message
}
