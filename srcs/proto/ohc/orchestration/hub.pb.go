// Code generated manually from hub.proto (edition 2024 opaque API stub).
// Provides the same public API as the generated protoc output.
package orchestrationpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Agent ───────────────────────────────────────────────────────────────────

type Agent struct {
	id             string
	name           string
	role           string
	organizationId string
	agentStatus    string
}

func (a *Agent) GetId() string             { return a.id }
func (a *Agent) GetName() string           { return a.name }
func (a *Agent) GetRole() string           { return a.role }
func (a *Agent) GetOrganizationId() string { return a.organizationId }
func (a *Agent) GetStatus() string         { return a.agentStatus }

type Agent_builder struct {
	Id             string
	Name           string
	Role           string
	OrganizationId string
	Status         string
}

func (b Agent_builder) Build() *Agent {
	return &Agent{id: b.Id, name: b.Name, role: b.Role, organizationId: b.OrganizationId, agentStatus: b.Status}
}

// ─── Message ─────────────────────────────────────────────────────────────────

type Message struct {
	id             string
	fromAgent      string
	toAgent        string
	msgType        string
	content        string
	meetingId      string
	occurredAtUnix int64
}

func (m *Message) GetId() string             { return m.id }
func (m *Message) GetFromAgent() string      { return m.fromAgent }
func (m *Message) GetToAgent() string        { return m.toAgent }
func (m *Message) GetType() string           { return m.msgType }
func (m *Message) GetContent() string        { return m.content }
func (m *Message) GetMeetingId() string      { return m.meetingId }
func (m *Message) GetOccurredAtUnix() int64  { return m.occurredAtUnix }

type Message_builder struct {
	Id             string
	FromAgent      string
	ToAgent        string
	Type           string
	Content        string
	MeetingId      string
	OccurredAtUnix int64
}

func (b Message_builder) Build() *Message {
	return &Message{
		id: b.Id, fromAgent: b.FromAgent, toAgent: b.ToAgent,
		msgType: b.Type, content: b.Content, meetingId: b.MeetingId,
		occurredAtUnix: b.OccurredAtUnix,
	}
}

// ─── MeetingRoom ─────────────────────────────────────────────────────────────

type MeetingRoom struct {
	id           string
	agenda       string
	participants []string
	transcript   []*Message
}

func (r *MeetingRoom) GetId() string              { return r.id }
func (r *MeetingRoom) GetAgenda() string          { return r.agenda }
func (r *MeetingRoom) GetParticipants() []string  { return r.participants }
func (r *MeetingRoom) GetTranscript() []*Message  { return r.transcript }

type MeetingRoom_builder struct {
	Id           string
	Agenda       string
	Participants []string
	Transcript   []*Message
}

func (b MeetingRoom_builder) Build() *MeetingRoom {
	return &MeetingRoom{id: b.Id, agenda: b.Agenda, participants: b.Participants, transcript: b.Transcript}
}

// ─── RegisterAgentRequest ────────────────────────────────────────────────────

type RegisterAgentRequest struct{ agent *Agent }

func (r *RegisterAgentRequest) GetAgent() *Agent { return r.agent }

type RegisterAgentRequest_builder struct{ Agent *Agent }

func (b RegisterAgentRequest_builder) Build() *RegisterAgentRequest {
	return &RegisterAgentRequest{agent: b.Agent}
}

// ─── RegisterAgentResponse ───────────────────────────────────────────────────

type RegisterAgentResponse struct{ success bool }

func (r *RegisterAgentResponse) GetSuccess() bool { return r.success }

type RegisterAgentResponse_builder struct{ Success bool }

func (b RegisterAgentResponse_builder) Build() *RegisterAgentResponse {
	return &RegisterAgentResponse{success: b.Success}
}

// ─── OpenMeetingRequest ──────────────────────────────────────────────────────

type OpenMeetingRequest struct {
	meetingId    string
	agenda       string
	participants []string
}

func (r *OpenMeetingRequest) GetMeetingId() string     { return r.meetingId }
func (r *OpenMeetingRequest) GetAgenda() string        { return r.agenda }
func (r *OpenMeetingRequest) GetParticipants() []string { return r.participants }

type OpenMeetingRequest_builder struct {
	MeetingId    string
	Agenda       string
	Participants []string
}

func (b OpenMeetingRequest_builder) Build() *OpenMeetingRequest {
	return &OpenMeetingRequest{meetingId: b.MeetingId, agenda: b.Agenda, participants: b.Participants}
}

// ─── PublishMessageRequest ───────────────────────────────────────────────────

type PublishMessageRequest struct{ message *Message }

func (r *PublishMessageRequest) GetMessage() *Message { return r.message }

type PublishMessageRequest_builder struct{ Message *Message }

func (b PublishMessageRequest_builder) Build() *PublishMessageRequest {
	return &PublishMessageRequest{message: b.Message}
}

// ─── PublishMessageResponse ──────────────────────────────────────────────────

type PublishMessageResponse struct{ success bool }

func (r *PublishMessageResponse) GetSuccess() bool { return r.success }

type PublishMessageResponse_builder struct{ Success bool }

func (b PublishMessageResponse_builder) Build() *PublishMessageResponse {
	return &PublishMessageResponse{success: b.Success}
}

// ─── StreamMessagesRequest ───────────────────────────────────────────────────

type StreamMessagesRequest struct{ agentId string }

func (r *StreamMessagesRequest) GetAgentId() string { return r.agentId }

type StreamMessagesRequest_builder struct{ AgentId string }

func (b StreamMessagesRequest_builder) Build() *StreamMessagesRequest {
	return &StreamMessagesRequest{agentId: b.AgentId}
}

// ─── ReasonRequest ───────────────────────────────────────────────────────────

type ReasonRequest struct{ prompt string }

func (r *ReasonRequest) GetPrompt() string { return r.prompt }

type ReasonRequest_builder struct{ Prompt string }

func (b ReasonRequest_builder) Build() *ReasonRequest {
	return &ReasonRequest{prompt: b.Prompt}
}

// ─── ReasonResponse ──────────────────────────────────────────────────────────

type ReasonResponse struct{ content string }

func (r *ReasonResponse) GetContent() string { return r.content }

type ReasonResponse_builder struct{ Content string }

func (b ReasonResponse_builder) Build() *ReasonResponse {
	return &ReasonResponse{content: b.Content}
}

// ─── gRPC service ────────────────────────────────────────────────────────────

// HubServiceServer is the server API for HubService service.
type HubServiceServer interface {
	RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error)
	OpenMeeting(context.Context, *OpenMeetingRequest) (*MeetingRoom, error)
	Publish(context.Context, *PublishMessageRequest) (*PublishMessageResponse, error)
	StreamMessages(*StreamMessagesRequest, HubService_StreamMessagesServer) error
	Reason(context.Context, *ReasonRequest) (*ReasonResponse, error)
	mustEmbedUnimplementedHubServiceServer()
}

// UnimplementedHubServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHubServiceServer struct{}

func (UnimplementedHubServiceServer) RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedHubServiceServer) OpenMeeting(context.Context, *OpenMeetingRequest) (*MeetingRoom, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OpenMeeting not implemented")
}
func (UnimplementedHubServiceServer) Publish(context.Context, *PublishMessageRequest) (*PublishMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Publish not implemented")
}
func (UnimplementedHubServiceServer) StreamMessages(*StreamMessagesRequest, HubService_StreamMessagesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamMessages not implemented")
}
func (UnimplementedHubServiceServer) Reason(context.Context, *ReasonRequest) (*ReasonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reason not implemented")
}
func (UnimplementedHubServiceServer) mustEmbedUnimplementedHubServiceServer() {}

// UnsafeHubServiceServer may be embedded to opt out of forward compatibility.
type UnsafeHubServiceServer interface {
	mustEmbedUnimplementedHubServiceServer()
}

// HubService_StreamMessagesServer is the server stream for the StreamMessages RPC.
type HubService_StreamMessagesServer interface {
	Send(*Message) error
	grpc.ServerStream
}

// RegisterHubServiceServer registers the HubServiceServer with the given gRPC server.
func RegisterHubServiceServer(s grpc.ServiceRegistrar, srv HubServiceServer) {
	s.RegisterService(&HubService_ServiceDesc, srv)
}

// HubService_ServiceDesc is the grpc.ServiceDesc for HubService service.
var HubService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ohc.orchestration.HubService",
	HandlerType: (*HubServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterAgent",
			Handler:    _HubService_RegisterAgent_Handler,
		},
		{
			MethodName: "OpenMeeting",
			Handler:    _HubService_OpenMeeting_Handler,
		},
		{
			MethodName: "Publish",
			Handler:    _HubService_Publish_Handler,
		},
		{
			MethodName: "Reason",
			Handler:    _HubService_Reason_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMessages",
			Handler:       _HubService_StreamMessages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "hub.proto",
}

func _HubService_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServiceServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/ohc.orchestration.HubService/RegisterAgent"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServiceServer).RegisterAgent(ctx, req.(*RegisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HubService_OpenMeeting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenMeetingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServiceServer).OpenMeeting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/ohc.orchestration.HubService/OpenMeeting"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServiceServer).OpenMeeting(ctx, req.(*OpenMeetingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HubService_Publish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServiceServer).Publish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/ohc.orchestration.HubService/Publish"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServiceServer).Publish(ctx, req.(*PublishMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HubService_Reason_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReasonRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServiceServer).Reason(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/ohc.orchestration.HubService/Reason"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServiceServer).Reason(ctx, req.(*ReasonRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type hubServiceStreamMessagesServer struct {
	grpc.ServerStream
}

func (x *hubServiceStreamMessagesServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _HubService_StreamMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamMessagesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HubServiceServer).StreamMessages(m, &hubServiceStreamMessagesServer{stream})
}
